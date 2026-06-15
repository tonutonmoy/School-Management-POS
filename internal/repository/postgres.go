package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository interface {
	GetByEmail(ctx context.Context, email string) (*UserRecord, error)
	GetByID(ctx context.Context, id uuid.UUID) (*UserRecord, error)
	Create(ctx context.Context, params CreateUserParams) (*UserRecord, error)
	Update(ctx context.Context, id uuid.UUID, params UpdateUserParams) (*UserRecord, error)
	SoftDelete(ctx context.Context, id uuid.UUID) error
	SetActive(ctx context.Context, id uuid.UUID, active bool) (*UserRecord, error)
	UpdatePassword(ctx context.Context, id uuid.UUID, hash string) error
	UpdateLastLogin(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int32) ([]UserRecord, error)
	Count(ctx context.Context) (int64, error)
	CountByRoleSlug(ctx context.Context, slug string) (int64, error)
	GetPermissions(ctx context.Context, userID uuid.UUID) ([]string, error)
}

type UserRecord struct {
	ID           uuid.UUID
	Email        string
	PasswordHash string
	FirstName    string
	LastName     string
	Phone        string
	RoleID       uuid.UUID
	RoleName     string
	RoleSlug     string
	IsActive     bool
	LastLoginAt  *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type CreateUserParams struct {
	Email        string
	PasswordHash string
	FirstName    string
	LastName     string
	Phone        string
	RoleID       uuid.UUID
	IsActive     bool
}

type UpdateUserParams struct {
	Email     string
	FirstName string
	LastName  string
	Phone     string
	RoleID    uuid.UUID
	IsActive  bool
}

type userRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) UserRepository {
	return &userRepository{pool: pool}
}

const userSelect = `
SELECT u.id, u.email, u.password_hash, u.first_name, u.last_name,
       COALESCE(u.phone, ''), u.role_id, u.is_active, u.last_login_at,
       u.created_at, u.updated_at, r.name, r.slug
FROM users u
JOIN roles r ON r.id = u.role_id`

func scanUser(row pgx.Row) (*UserRecord, error) {
	var u UserRecord
	var lastLogin pgtype.Timestamptz
	if err := row.Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.FirstName, &u.LastName, &u.Phone,
		&u.RoleID, &u.IsActive, &lastLogin, &u.CreatedAt, &u.UpdatedAt,
		&u.RoleName, &u.RoleSlug,
	); err != nil {
		return nil, err
	}
	if lastLogin.Valid {
		u.LastLoginAt = &lastLogin.Time
	}
	return &u, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*UserRecord, error) {
	q := userSelect + ` WHERE LOWER(u.email) = LOWER($1) AND u.deleted_at IS NULL`
	row := r.pool.QueryRow(ctx, q, email)
	u, err := scanUser(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return u, nil
}

func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*UserRecord, error) {
	q := userSelect + ` WHERE u.id = $1 AND u.deleted_at IS NULL`
	row := r.pool.QueryRow(ctx, q, id)
	u, err := scanUser(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return u, nil
}

func (r *userRepository) Create(ctx context.Context, params CreateUserParams) (*UserRecord, error) {
	var id uuid.UUID
	err := r.pool.QueryRow(ctx, `
INSERT INTO users (email, password_hash, first_name, last_name, phone, role_id, is_active)
VALUES ($1, $2, $3, $4, NULLIF($5, ''), $6, $7)
RETURNING id`, params.Email, params.PasswordHash, params.FirstName, params.LastName,
		params.Phone, params.RoleID, params.IsActive).Scan(&id)
	if err != nil {
		return nil, err
	}
	return r.GetByID(ctx, id)
}

func (r *userRepository) Update(ctx context.Context, id uuid.UUID, params UpdateUserParams) (*UserRecord, error) {
	tag, err := r.pool.Exec(ctx, `
UPDATE users SET email=$2, first_name=$3, last_name=$4, phone=NULLIF($5, ''),
role_id=$6, is_active=$7, updated_at=NOW()
WHERE id=$1 AND deleted_at IS NULL`,
		id, params.Email, params.FirstName, params.LastName, params.Phone, params.RoleID, params.IsActive)
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		return nil, nil
	}
	return r.GetByID(ctx, id)
}

func (r *userRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE users SET deleted_at=NOW(), updated_at=NOW() WHERE id=$1 AND deleted_at IS NULL`, id)
	return err
}

func (r *userRepository) SetActive(ctx context.Context, id uuid.UUID, active bool) (*UserRecord, error) {
	tag, err := r.pool.Exec(ctx, `UPDATE users SET is_active=$2, updated_at=NOW() WHERE id=$1 AND deleted_at IS NULL`, id, active)
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		return nil, nil
	}
	return r.GetByID(ctx, id)
}

func (r *userRepository) UpdatePassword(ctx context.Context, id uuid.UUID, hash string) error {
	_, err := r.pool.Exec(ctx, `UPDATE users SET password_hash=$2, updated_at=NOW() WHERE id=$1 AND deleted_at IS NULL`, id, hash)
	return err
}

func (r *userRepository) UpdateLastLogin(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE users SET last_login_at=NOW(), updated_at=NOW() WHERE id=$1`, id)
	return err
}

func (r *userRepository) List(ctx context.Context, limit, offset int32) ([]UserRecord, error) {
	q := userSelect + ` WHERE u.deleted_at IS NULL ORDER BY u.created_at DESC LIMIT $1 OFFSET $2`
	rows, err := r.pool.Query(ctx, q, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []UserRecord
	for rows.Next() {
		var u UserRecord
		var lastLogin pgtype.Timestamptz
		if err := rows.Scan(
			&u.ID, &u.Email, &u.PasswordHash, &u.FirstName, &u.LastName, &u.Phone,
			&u.RoleID, &u.IsActive, &lastLogin, &u.CreatedAt, &u.UpdatedAt,
			&u.RoleName, &u.RoleSlug,
		); err != nil {
			return nil, err
		}
		if lastLogin.Valid {
			u.LastLoginAt = &lastLogin.Time
		}
		items = append(items, u)
	}
	return items, rows.Err()
}

func (r *userRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE deleted_at IS NULL`).Scan(&count)
	return count, err
}

func (r *userRepository) CountByRoleSlug(ctx context.Context, slug string) (int64, error) {
	var count int64
	err := r.pool.QueryRow(ctx, `
SELECT COUNT(*) FROM users u JOIN roles r ON r.id=u.role_id
WHERE u.deleted_at IS NULL AND r.slug=$1`, slug).Scan(&count)
	return count, err
}

func (r *userRepository) GetPermissions(ctx context.Context, userID uuid.UUID) ([]string, error) {
	rows, err := r.pool.Query(ctx, `
SELECT p.slug FROM permissions p
JOIN role_permissions rp ON rp.permission_id = p.id
JOIN users u ON u.role_id = rp.role_id
WHERE u.id = $1`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var slugs []string
	for rows.Next() {
		var slug string
		if err := rows.Scan(&slug); err != nil {
			return nil, err
		}
		slugs = append(slugs, slug)
	}
	return slugs, rows.Err()
}

type RoleRepository interface {
	Create(ctx context.Context, name, slug, description string, isSystem bool) (*RoleRecord, error)
	GetByID(ctx context.Context, id uuid.UUID) (*RoleRecord, error)
	GetBySlug(ctx context.Context, slug string) (*RoleRecord, error)
	Update(ctx context.Context, id uuid.UUID, name, slug, description string) (*RoleRecord, error)
	SoftDelete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context) ([]RoleRecord, error)
	AssignPermission(ctx context.Context, roleID, permissionID uuid.UUID) error
	ClearPermissions(ctx context.Context, roleID uuid.UUID) error
	GetPermissions(ctx context.Context, roleID uuid.UUID) ([]PermissionRecord, error)
	GetPermissionSlugs(ctx context.Context, roleID uuid.UUID) ([]string, error)
}

type RoleRecord struct {
	ID          uuid.UUID
	Name        string
	Slug        string
	Description string
	IsSystem    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type roleRepository struct {
	pool *pgxpool.Pool
}

func NewRoleRepository(pool *pgxpool.Pool) RoleRepository {
	return &roleRepository{pool: pool}
}

func scanRole(row pgx.Row) (*RoleRecord, error) {
	var role RoleRecord
	var desc pgtype.Text
	if err := row.Scan(&role.ID, &role.Name, &role.Slug, &desc, &role.IsSystem, &role.CreatedAt, &role.UpdatedAt); err != nil {
		return nil, err
	}
	if desc.Valid {
		role.Description = desc.String
	}
	return &role, nil
}

func (r *roleRepository) Create(ctx context.Context, name, slug, description string, isSystem bool) (*RoleRecord, error) {
	var id uuid.UUID
	err := r.pool.QueryRow(ctx, `
INSERT INTO roles (name, slug, description, is_system) VALUES ($1,$2,NULLIF($3,''),$4) RETURNING id`,
		name, slug, description, isSystem).Scan(&id)
	if err != nil {
		return nil, err
	}
	return r.GetByID(ctx, id)
}

func (r *roleRepository) GetByID(ctx context.Context, id uuid.UUID) (*RoleRecord, error) {
	row := r.pool.QueryRow(ctx, `SELECT id,name,slug,description,is_system,created_at,updated_at FROM roles WHERE id=$1 AND deleted_at IS NULL`, id)
	role, err := scanRole(row)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return role, err
}

func (r *roleRepository) GetBySlug(ctx context.Context, slug string) (*RoleRecord, error) {
	row := r.pool.QueryRow(ctx, `SELECT id,name,slug,description,is_system,created_at,updated_at FROM roles WHERE slug=$1 AND deleted_at IS NULL`, slug)
	role, err := scanRole(row)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return role, err
}

func (r *roleRepository) Update(ctx context.Context, id uuid.UUID, name, slug, description string) (*RoleRecord, error) {
	tag, err := r.pool.Exec(ctx, `
UPDATE roles SET name=$2, slug=$3, description=NULLIF($4,''), updated_at=NOW()
WHERE id=$1 AND deleted_at IS NULL AND is_system=FALSE`, id, name, slug, description)
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		return nil, nil
	}
	return r.GetByID(ctx, id)
}

func (r *roleRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE roles SET deleted_at=NOW(), updated_at=NOW() WHERE id=$1 AND deleted_at IS NULL AND is_system=FALSE`, id)
	return err
}

func (r *roleRepository) List(ctx context.Context) ([]RoleRecord, error) {
	rows, err := r.pool.Query(ctx, `SELECT id,name,slug,description,is_system,created_at,updated_at FROM roles WHERE deleted_at IS NULL ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []RoleRecord
	for rows.Next() {
		var role RoleRecord
		var desc pgtype.Text
		if err := rows.Scan(&role.ID, &role.Name, &role.Slug, &desc, &role.IsSystem, &role.CreatedAt, &role.UpdatedAt); err != nil {
			return nil, err
		}
		if desc.Valid {
			role.Description = desc.String
		}
		items = append(items, role)
	}
	return items, rows.Err()
}

func (r *roleRepository) AssignPermission(ctx context.Context, roleID, permissionID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO role_permissions (role_id, permission_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`, roleID, permissionID)
	return err
}

func (r *roleRepository) ClearPermissions(ctx context.Context, roleID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM role_permissions WHERE role_id=$1`, roleID)
	return err
}

func (r *roleRepository) GetPermissions(ctx context.Context, roleID uuid.UUID) ([]PermissionRecord, error) {
	rows, err := r.pool.Query(ctx, `
SELECT p.id,p.name,p.slug,p.description,p.module,p.created_at,p.updated_at
FROM permissions p JOIN role_permissions rp ON rp.permission_id=p.id
WHERE rp.role_id=$1 ORDER BY p.module,p.name`, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanPermissions(rows)
}

func (r *roleRepository) GetPermissionSlugs(ctx context.Context, roleID uuid.UUID) ([]string, error) {
	rows, err := r.pool.Query(ctx, `
SELECT p.slug FROM permissions p JOIN role_permissions rp ON rp.permission_id=p.id WHERE rp.role_id=$1`, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var slugs []string
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			return nil, err
		}
		slugs = append(slugs, s)
	}
	return slugs, rows.Err()
}

type PermissionRepository interface {
	List(ctx context.Context) ([]PermissionRecord, error)
	GetByID(ctx context.Context, id uuid.UUID) (*PermissionRecord, error)
	GetBySlug(ctx context.Context, slug string) (*PermissionRecord, error)
	Create(ctx context.Context, name, slug, description, module string) (*PermissionRecord, error)
	Update(ctx context.Context, id uuid.UUID, name, slug, description, module string) (*PermissionRecord, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type PermissionRecord struct {
	ID          uuid.UUID
	Name        string
	Slug        string
	Description string
	Module      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type permissionRepository struct {
	pool *pgxpool.Pool
}

func NewPermissionRepository(pool *pgxpool.Pool) PermissionRepository {
	return &permissionRepository{pool: pool}
}

func scanPermission(row pgx.Row) (*PermissionRecord, error) {
	var p PermissionRecord
	var desc pgtype.Text
	if err := row.Scan(&p.ID, &p.Name, &p.Slug, &desc, &p.Module, &p.CreatedAt, &p.UpdatedAt); err != nil {
		return nil, err
	}
	if desc.Valid {
		p.Description = desc.String
	}
	return &p, nil
}

func scanPermissions(rows pgx.Rows) ([]PermissionRecord, error) {
	var items []PermissionRecord
	for rows.Next() {
		var p PermissionRecord
		var desc pgtype.Text
		if err := rows.Scan(&p.ID, &p.Name, &p.Slug, &desc, &p.Module, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		if desc.Valid {
			p.Description = desc.String
		}
		items = append(items, p)
	}
	return items, rows.Err()
}

func (r *permissionRepository) List(ctx context.Context) ([]PermissionRecord, error) {
	rows, err := r.pool.Query(ctx, `SELECT id,name,slug,description,module,created_at,updated_at FROM permissions ORDER BY module,name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanPermissions(rows)
}

func (r *permissionRepository) GetByID(ctx context.Context, id uuid.UUID) (*PermissionRecord, error) {
	row := r.pool.QueryRow(ctx, `SELECT id,name,slug,description,module,created_at,updated_at FROM permissions WHERE id=$1`, id)
	p, err := scanPermission(row)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return p, err
}

func (r *permissionRepository) GetBySlug(ctx context.Context, slug string) (*PermissionRecord, error) {
	row := r.pool.QueryRow(ctx, `SELECT id,name,slug,description,module,created_at,updated_at FROM permissions WHERE slug=$1`, slug)
	p, err := scanPermission(row)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return p, err
}

func (r *permissionRepository) Create(ctx context.Context, name, slug, description, module string) (*PermissionRecord, error) {
	var id uuid.UUID
	err := r.pool.QueryRow(ctx, `
INSERT INTO permissions (name,slug,description,module) VALUES ($1,$2,NULLIF($3,''),$4) RETURNING id`,
		name, slug, description, module).Scan(&id)
	if err != nil {
		return nil, err
	}
	return r.GetByID(ctx, id)
}

func (r *permissionRepository) Update(ctx context.Context, id uuid.UUID, name, slug, description, module string) (*PermissionRecord, error) {
	tag, err := r.pool.Exec(ctx, `
UPDATE permissions SET name=$2,slug=$3,description=NULLIF($4,''),module=$5,updated_at=NOW() WHERE id=$1`,
		id, name, slug, description, module)
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		return nil, nil
	}
	return r.GetByID(ctx, id)
}

func (r *permissionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM permissions WHERE id=$1`, id)
	return err
}

type AuthRepository interface {
	CreatePasswordResetToken(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) error
	GetValidPasswordResetToken(ctx context.Context, tokenHash string) (*PasswordResetRecord, error)
	MarkPasswordResetTokenUsed(ctx context.Context, id uuid.UUID) error
	RevokeToken(ctx context.Context, jti string, expiresAt time.Time) error
	IsTokenRevoked(ctx context.Context, jti string) (bool, error)
}

type PasswordResetRecord struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	TokenHash string
	ExpiresAt time.Time
}

type authRepository struct {
	pool *pgxpool.Pool
}

func NewAuthRepository(pool *pgxpool.Pool) AuthRepository {
	return &authRepository{pool: pool}
}

func (r *authRepository) CreatePasswordResetToken(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) error {
	_, err := r.pool.Exec(ctx, `
INSERT INTO password_reset_tokens (user_id, token_hash, expires_at) VALUES ($1,$2,$3)`, userID, tokenHash, expiresAt)
	return err
}

func (r *authRepository) GetValidPasswordResetToken(ctx context.Context, tokenHash string) (*PasswordResetRecord, error) {
	row := r.pool.QueryRow(ctx, `
SELECT id,user_id,token_hash,expires_at FROM password_reset_tokens
WHERE token_hash=$1 AND used_at IS NULL AND expires_at > NOW()
ORDER BY created_at DESC LIMIT 1`, tokenHash)
	var rec PasswordResetRecord
	if err := row.Scan(&rec.ID, &rec.UserID, &rec.TokenHash, &rec.ExpiresAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &rec, nil
}

func (r *authRepository) MarkPasswordResetTokenUsed(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE password_reset_tokens SET used_at=NOW() WHERE id=$1`, id)
	return err
}

func (r *authRepository) RevokeToken(ctx context.Context, jti string, expiresAt time.Time) error {
	_, err := r.pool.Exec(ctx, `
INSERT INTO revoked_tokens (jti, expires_at) VALUES ($1,$2) ON CONFLICT (jti) DO NOTHING`, jti, expiresAt)
	return err
}

func (r *authRepository) IsTokenRevoked(ctx context.Context, jti string) (bool, error) {
	var revoked bool
	err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM revoked_tokens WHERE jti=$1)`, jti).Scan(&revoked)
	return revoked, err
}

type SchoolRepository interface {
	Get(ctx context.Context) (*SchoolRecord, error)
	Create(ctx context.Context, params SchoolParams) (*SchoolRecord, error)
	Update(ctx context.Context, id uuid.UUID, params SchoolParams) (*SchoolRecord, error)
}

type SchoolRecord struct {
	ID      uuid.UUID
	Name    string
	LogoURL string
	Address string
	Email   string
	Phone   string
	Website string
}

type SchoolParams struct {
	Name    string
	LogoURL string
	Address string
	Email   string
	Phone   string
	Website string
}

type schoolRepository struct {
	pool *pgxpool.Pool
}

func NewSchoolRepository(pool *pgxpool.Pool) SchoolRepository {
	return &schoolRepository{pool: pool}
}

func scanSchool(row pgx.Row) (*SchoolRecord, error) {
	var s SchoolRecord
	var logo, addr, email, phone, website pgtype.Text
	var createdAt, updatedAt time.Time
	if err := row.Scan(&s.ID, &s.Name, &logo, &addr, &email, &phone, &website, &createdAt, &updatedAt); err != nil {
		return nil, err
	}
	if logo.Valid {
		s.LogoURL = logo.String
	}
	if addr.Valid {
		s.Address = addr.String
	}
	if email.Valid {
		s.Email = email.String
	}
	if phone.Valid {
		s.Phone = phone.String
	}
	if website.Valid {
		s.Website = website.String
	}
	return &s, nil
}

func (r *schoolRepository) Get(ctx context.Context) (*SchoolRecord, error) {
	row := r.pool.QueryRow(ctx, `SELECT id,name,logo_url,address,email,phone,website,created_at,updated_at FROM schools ORDER BY created_at ASC LIMIT 1`)
	s, err := scanSchool(row)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return s, err
}

func (r *schoolRepository) Create(ctx context.Context, params SchoolParams) (*SchoolRecord, error) {
	var id uuid.UUID
	err := r.pool.QueryRow(ctx, `
INSERT INTO schools (name,logo_url,address,email,phone,website)
VALUES ($1,NULLIF($2,''),NULLIF($3,''),NULLIF($4,''),NULLIF($5,''),NULLIF($6,'')) RETURNING id`,
		params.Name, params.LogoURL, params.Address, params.Email, params.Phone, params.Website).Scan(&id)
	if err != nil {
		return nil, err
	}
	return r.Get(ctx)
}

func (r *schoolRepository) Update(ctx context.Context, id uuid.UUID, params SchoolParams) (*SchoolRecord, error) {
	tag, err := r.pool.Exec(ctx, `
UPDATE schools SET name=$2,
logo_url=CASE WHEN $3='' THEN logo_url ELSE $3 END,
address=NULLIF($4,''),email=NULLIF($5,''),phone=NULLIF($6,''),website=NULLIF($7,''),updated_at=NOW()
WHERE id=$1`, id, params.Name, params.LogoURL, params.Address, params.Email, params.Phone, params.Website)
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		return nil, nil
	}
	row := r.pool.QueryRow(ctx, `SELECT id,name,logo_url,address,email,phone,website,created_at,updated_at FROM schools WHERE id=$1`, id)
	return scanSchool(row)
}

type AcademicSessionRepository interface {
	Create(ctx context.Context, name string, start, end time.Time, active bool) (*SessionRecord, error)
	GetByID(ctx context.Context, id uuid.UUID) (*SessionRecord, error)
	Update(ctx context.Context, id uuid.UUID, name string, start, end time.Time, active bool) (*SessionRecord, error)
	SoftDelete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context) ([]SessionRecord, error)
	GetActive(ctx context.Context) (*SessionRecord, error)
	SetActive(ctx context.Context, id uuid.UUID) (*SessionRecord, error)
	DeactivateAll(ctx context.Context) error
}

type SessionRecord struct {
	ID        uuid.UUID
	Name      string
	StartDate time.Time
	EndDate   time.Time
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

type academicSessionRepository struct {
	pool *pgxpool.Pool
}

func NewAcademicSessionRepository(pool *pgxpool.Pool) AcademicSessionRepository {
	return &academicSessionRepository{pool: pool}
}

func scanSession(row pgx.Row) (*SessionRecord, error) {
	var s SessionRecord
	if err := row.Scan(&s.ID, &s.Name, &s.StartDate, &s.EndDate, &s.IsActive, &s.CreatedAt, &s.UpdatedAt); err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *academicSessionRepository) Create(ctx context.Context, name string, start, end time.Time, active bool) (*SessionRecord, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	if active {
		if _, err := tx.Exec(ctx, `UPDATE academic_sessions SET is_active=FALSE, updated_at=NOW() WHERE deleted_at IS NULL AND is_active=TRUE`); err != nil {
			return nil, err
		}
	}

	var id uuid.UUID
	if err := tx.QueryRow(ctx, `
INSERT INTO academic_sessions (name,start_date,end_date,is_active) VALUES ($1,$2,$3,$4) RETURNING id`,
		name, start, end, active).Scan(&id); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return r.GetByID(ctx, id)
}

func (r *academicSessionRepository) GetByID(ctx context.Context, id uuid.UUID) (*SessionRecord, error) {
	row := r.pool.QueryRow(ctx, `
SELECT id,name,start_date,end_date,is_active,created_at,updated_at FROM academic_sessions
WHERE id=$1 AND deleted_at IS NULL`, id)
	s, err := scanSession(row)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return s, err
}

func (r *academicSessionRepository) Update(ctx context.Context, id uuid.UUID, name string, start, end time.Time, active bool) (*SessionRecord, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	if active {
		if _, err := tx.Exec(ctx, `UPDATE academic_sessions SET is_active=FALSE, updated_at=NOW() WHERE deleted_at IS NULL AND is_active=TRUE AND id<>$1`, id); err != nil {
			return nil, err
		}
	}

	tag, err := tx.Exec(ctx, `
UPDATE academic_sessions SET name=$2,start_date=$3,end_date=$4,is_active=$5,updated_at=NOW()
WHERE id=$1 AND deleted_at IS NULL`, id, name, start, end, active)
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		return nil, nil
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return r.GetByID(ctx, id)
}

func (r *academicSessionRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE academic_sessions SET deleted_at=NOW(), updated_at=NOW(), is_active=FALSE WHERE id=$1 AND deleted_at IS NULL`, id)
	return err
}

func (r *academicSessionRepository) List(ctx context.Context) ([]SessionRecord, error) {
	rows, err := r.pool.Query(ctx, `
SELECT id,name,start_date,end_date,is_active,created_at,updated_at FROM academic_sessions
WHERE deleted_at IS NULL ORDER BY start_date DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []SessionRecord
	for rows.Next() {
		var s SessionRecord
		if err := rows.Scan(&s.ID, &s.Name, &s.StartDate, &s.EndDate, &s.IsActive, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, s)
	}
	return items, rows.Err()
}

func (r *academicSessionRepository) GetActive(ctx context.Context) (*SessionRecord, error) {
	row := r.pool.QueryRow(ctx, `
SELECT id,name,start_date,end_date,is_active,created_at,updated_at FROM academic_sessions
WHERE is_active=TRUE AND deleted_at IS NULL LIMIT 1`)
	s, err := scanSession(row)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return s, err
}

func (r *academicSessionRepository) SetActive(ctx context.Context, id uuid.UUID) (*SessionRecord, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `UPDATE academic_sessions SET is_active=FALSE, updated_at=NOW() WHERE deleted_at IS NULL AND is_active=TRUE`); err != nil {
		return nil, err
	}
	tag, err := tx.Exec(ctx, `UPDATE academic_sessions SET is_active=TRUE, updated_at=NOW() WHERE id=$1 AND deleted_at IS NULL`, id)
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		return nil, nil
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return r.GetByID(ctx, id)
}

func (r *academicSessionRepository) DeactivateAll(ctx context.Context) error {
	_, err := r.pool.Exec(ctx, `UPDATE academic_sessions SET is_active=FALSE, updated_at=NOW() WHERE deleted_at IS NULL AND is_active=TRUE`)
	return err
}

type AuditRepository interface {
	Create(ctx context.Context, params AuditParams) error
	ListRecent(ctx context.Context, limit int32) ([]AuditRecord, error)
	List(ctx context.Context, limit, offset int32) ([]AuditRecord, error)
	Count(ctx context.Context) (int64, error)
	Search(ctx context.Context, f AuditSearchParams) ([]AuditRecord, error)
	CountSearch(ctx context.Context, f AuditSearchParams) (int64, error)
}

type AuditParams struct {
	UserID     *uuid.UUID
	Action     string
	EntityType string
	EntityID   *uuid.UUID
	IPAddress  string
	Metadata   map[string]any
}

type AuditRecord struct {
	ID          uuid.UUID
	UserID      *uuid.UUID
	Action      string
	EntityType  string
	EntityID    *uuid.UUID
	IPAddress   string
	UserEmail   string
	UserName    string
	Description string
	CreatedAt   time.Time
}

type auditRepository struct {
	pool *pgxpool.Pool
}

func NewAuditRepository(pool *pgxpool.Pool) AuditRepository {
	return &auditRepository{pool: pool}
}

func (r *auditRepository) Create(ctx context.Context, params AuditParams) error {
	meta, err := json.Marshal(params.Metadata)
	if err != nil {
		meta = []byte("{}")
	}
	var userID pgtype.UUID
	if params.UserID != nil {
		userID = pgtype.UUID{Bytes: *params.UserID, Valid: true}
	}
	var entityID pgtype.UUID
	if params.EntityID != nil {
		entityID = pgtype.UUID{Bytes: *params.EntityID, Valid: true}
	}
	var ip pgtype.Text
	if params.IPAddress != "" {
		if parsed := net.ParseIP(params.IPAddress); parsed != nil {
			ip = pgtype.Text{String: params.IPAddress, Valid: true}
		}
	}
	_, err = r.pool.Exec(ctx, `
INSERT INTO audit_logs (user_id, action, entity_type, entity_id, ip_address, metadata)
VALUES ($1,$2,$3,$4,$5,$6)`, userID, params.Action, params.EntityType, entityID, ip, meta)
	return err
}

func scanAudit(rows pgx.Rows) ([]AuditRecord, error) {
	var items []AuditRecord
	for rows.Next() {
		var rec AuditRecord
		var userID, entityID pgtype.UUID
		var ip pgtype.Text
		var email, first, last pgtype.Text
		if err := rows.Scan(
			&rec.ID, &userID, &rec.Action, &rec.EntityType, &entityID, &ip, &rec.CreatedAt,
			&email, &first, &last,
		); err != nil {
			return nil, err
		}
		if userID.Valid {
			id := uuid.UUID(userID.Bytes)
			rec.UserID = &id
		}
		if entityID.Valid {
			id := uuid.UUID(entityID.Bytes)
			rec.EntityID = &id
		}
		if ip.Valid {
			rec.IPAddress = ip.String
		}
		if email.Valid {
			rec.UserEmail = email.String
		}
		if first.Valid || last.Valid {
			rec.UserName = fmt.Sprintf("%s %s", first.String, last.String)
		}
		rec.Description = formatAuditDescription(rec)
		items = append(items, rec)
	}
	return items, rows.Err()
}

func formatAuditDescription(rec AuditRecord) string {
	who := rec.UserEmail
	if who == "" {
		who = "System"
	}
	return fmt.Sprintf("%s performed %s on %s", who, rec.Action, rec.EntityType)
}

func (r *auditRepository) ListRecent(ctx context.Context, limit int32) ([]AuditRecord, error) {
	rows, err := r.pool.Query(ctx, `
SELECT a.id,a.user_id,a.action,a.entity_type,a.entity_id,a.ip_address,a.created_at,
       u.email,u.first_name,u.last_name
FROM audit_logs a LEFT JOIN users u ON u.id=a.user_id
ORDER BY a.created_at DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanAudit(rows)
}

func (r *auditRepository) List(ctx context.Context, limit, offset int32) ([]AuditRecord, error) {
	rows, err := r.pool.Query(ctx, `
SELECT a.id,a.user_id,a.action,a.entity_type,a.entity_id,a.ip_address,a.created_at,
       u.email,u.first_name,u.last_name
FROM audit_logs a LEFT JOIN users u ON u.id=a.user_id
ORDER BY a.created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanAudit(rows)
}

func (r *auditRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM audit_logs`).Scan(&count)
	return count, err
}
