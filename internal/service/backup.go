package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/school-management/pos/internal/config"
	"github.com/school-management/pos/internal/dto"
	"github.com/school-management/pos/internal/model"
	"github.com/school-management/pos/internal/repository"
)

type BackupService struct {
	repos  *repository.Repositories
	cfg    *config.Config
	audit  *AuditService
	logger *slog.Logger
}

func NewBackupService(repos *repository.Repositories, cfg *config.Config, audit *AuditService, logger *slog.Logger) *BackupService {
	return &BackupService{repos: repos, cfg: cfg, audit: audit, logger: logger}
}

func (s *BackupService) EnsureDir() error {
	return os.MkdirAll(s.cfg.Backup.Dir, 0o750)
}

func (s *BackupService) CreateBackup(ctx context.Context, backupType string, actorID *uuid.UUID, ip string) (*dto.SystemBackupResponse, error) {
	if err := s.EnsureDir(); err != nil {
		return nil, err
	}
	fileName := fmt.Sprintf("backup_%s.sql", time.Now().Format("20060102_150405"))
	filePath := filepath.Join(s.cfg.Backup.Dir, fileName)
	rec, err := s.repos.System.CreateBackup(ctx, repository.CreateBackupParams{
		FileName: fileName, FilePath: filePath, BackupType: backupType, Status: model.BackupPending, CreatedBy: actorID,
	})
	if err != nil {
		return nil, err
	}
	if err := runPgDump(ctx, s.cfg.Database.URL, filePath); err != nil {
		_ = s.repos.System.UpdateBackup(ctx, rec.ID, repository.UpdateBackupParams{
			Status: model.BackupFailed, ErrorMessage: err.Error(),
		})
		s.audit.Log(ctx, actorID, model.ActionCreate, model.EntitySystemBackup, &rec.ID, ip, map[string]any{"status": "failed"})
		return nil, err
	}
	info, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}
	checksum, verified := fileChecksum(filePath)
	_ = s.repos.System.UpdateBackup(ctx, rec.ID, repository.UpdateBackupParams{
		Status: model.BackupCompleted, Checksum: checksum, FileSize: info.Size(), Verified: verified,
	})
	updated, _ := s.repos.System.GetBackup(ctx, rec.ID)
	resp := mapBackup(updated)
	s.audit.Log(ctx, actorID, model.ActionCreate, model.EntitySystemBackup, &rec.ID, ip, map[string]any{
		"file": fileName, "size": info.Size(), "verified": verified,
	})
	return &resp, nil
}

func (s *BackupService) RestoreBackup(ctx context.Context, id uuid.UUID, actorID uuid.UUID, ip string) error {
	rec, err := s.repos.System.GetBackup(ctx, id)
	if err != nil || rec == nil {
		return ErrNotFound
	}
	if rec.Status != model.BackupCompleted {
		return fmt.Errorf("%w: backup not completed", ErrValidation)
	}
	if !rec.Verified {
		return fmt.Errorf("%w: backup failed verification", ErrValidation)
	}
	if err := runPsqlRestore(ctx, s.cfg.Database.URL, rec.FilePath); err != nil {
		return err
	}
	s.audit.Log(ctx, &actorID, model.ActionUpdate, model.EntitySystemBackup, &id, ip, map[string]any{"action": "restore"})
	return nil
}

func (s *BackupService) VerifyBackup(ctx context.Context, id uuid.UUID) (bool, error) {
	rec, err := s.repos.System.GetBackup(ctx, id)
	if err != nil || rec == nil {
		return false, ErrNotFound
	}
	checksum, ok := fileChecksum(rec.FilePath)
	verified := ok && (rec.Checksum == "" || rec.Checksum == checksum)
	_ = s.repos.System.UpdateBackup(ctx, id, repository.UpdateBackupParams{
		Checksum: checksum, Verified: verified,
	})
	return verified, nil
}

func (s *BackupService) GetBackupFilePath(ctx context.Context, id uuid.UUID) (string, string, error) {
	rec, err := s.repos.System.GetBackup(ctx, id)
	if err != nil || rec == nil {
		return "", "", ErrNotFound
	}
	if rec.Status != model.BackupCompleted {
		return "", "", fmt.Errorf("%w: backup not available", ErrValidation)
	}
	return rec.FilePath, rec.FileName, nil
}

func (s *BackupService) ListBackups(ctx context.Context, page, pageSize int) (*dto.PaginatedBackups, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	total, err := s.repos.System.CountBackups(ctx)
	if err != nil {
		return nil, err
	}
	recs, err := s.repos.System.ListBackups(ctx, int32(pageSize), int32((page-1)*pageSize))
	if err != nil {
		return nil, err
	}
	items := make([]dto.SystemBackupResponse, 0, len(recs))
	for _, r := range recs {
		items = append(items, mapBackup(&r))
	}
	pages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		pages++
	}
	return &dto.PaginatedBackups{Items: items, Total: total, Page: page, PageSize: pageSize, TotalPages: pages}, nil
}

func (s *BackupService) RunScheduledBackup(ctx context.Context) error {
	settings, _ := s.repos.System.GetSetting(ctx, model.SettingCategoryGeneral, "backup_schedule")
	if enabled, _ := settings["enabled"].(bool); !enabled {
		return nil
	}
	_, err := s.CreateBackup(ctx, model.BackupScheduled, nil, "scheduler")
	if err != nil {
		s.logger.Warn("scheduled backup failed", "error", err)
		return err
	}
	if days, ok := settings["retention_days"].(float64); ok && days > 0 {
		before := time.Now().AddDate(0, 0, -int(days))
		deleted, _ := s.repos.System.DeleteOldBackups(ctx, before)
		s.logger.Info("backup retention cleanup", "deleted", deleted)
	}
	return nil
}

func (s *BackupService) StartScheduler(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				_ = s.RunScheduledBackup(context.Background())
			}
		}
	}()
}

func runPgDump(ctx context.Context, dbURL, outPath string) error {
	if _, err := exec.LookPath("pg_dump"); err != nil {
		return fmt.Errorf("pg_dump not found: install postgresql-client")
	}
	cmd := exec.CommandContext(ctx, "pg_dump", "--dbname", dbURL, "--file", outPath, "--format=plain", "--no-owner", "--no-acl")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("pg_dump failed: %s: %w", strings.TrimSpace(string(out)), err)
	}
	return nil
}

func runPsqlRestore(ctx context.Context, dbURL, filePath string) error {
	if _, err := exec.LookPath("psql"); err != nil {
		return fmt.Errorf("psql not found: install postgresql-client")
	}
	cmd := exec.CommandContext(ctx, "psql", "--dbname", dbURL, "--file", filePath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("psql restore failed: %s: %w", strings.TrimSpace(string(out)), err)
	}
	return nil
}

func fileChecksum(path string) (string, bool) {
	f, err := os.Open(path)
	if err != nil {
		return "", false
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", false
	}
	return hex.EncodeToString(h.Sum(nil)), true
}

func mapBackup(r *repository.BackupRecord) dto.SystemBackupResponse {
	if r == nil {
		return dto.SystemBackupResponse{}
	}
	return dto.SystemBackupResponse{
		ID: r.ID, FileName: r.FileName, FileSize: r.FileSize, BackupType: r.BackupType,
		Status: r.Status, Checksum: r.Checksum, Verified: r.Verified, ErrorMessage: r.ErrorMessage,
		CreatedByName: r.CreatedByName, CreatedAt: r.CreatedAt,
	}
}

func (s *BackupService) DiskUsageMB() float64 {
	var total int64
	_ = filepath.Walk(s.cfg.Backup.Dir, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			total += info.Size()
		}
		return nil
	})
	return float64(total) / (1024 * 1024)
}
