package storage

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"

	"github.com/school-management/pos/internal/config"
)

type Storage interface {
	Upload(ctx context.Context, filename string, body io.Reader, contentType string) (string, error)
	Enabled() bool
}

type R2Storage struct {
	client    *s3.Client
	bucket    string
	publicURL string
}

func New(cfg config.R2Config) (Storage, error) {
	if !cfg.Enabled {
		return &LocalStorage{baseURL: "/uploads"}, nil
	}

	resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...any) (aws.Endpoint, error) {
		return aws.Endpoint{URL: cfg.Endpoint, HostnameImmutable: true}, nil
	})

	awsCfg, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, "")),
		awsconfig.WithRegion("auto"),
		awsconfig.WithEndpointResolverWithOptions(resolver),
	)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	return &R2Storage{
		client:    client,
		bucket:    cfg.BucketName,
		publicURL: strings.TrimRight(cfg.PublicURL, "/"),
	}, nil
}

func (s *R2Storage) Enabled() bool { return true }

func (s *R2Storage) Upload(ctx context.Context, filename string, body io.Reader, contentType string) (string, error) {
	ext := filepath.Ext(filename)
	key := fmt.Sprintf("school-logos/%s-%d%s", uuid.NewString(), time.Now().Unix(), ext)
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        body,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", fmt.Errorf("upload to r2: %w", err)
	}
	if s.publicURL != "" {
		return fmt.Sprintf("%s/%s", s.publicURL, key), nil
	}
	return key, nil
}

type LocalStorage struct {
	baseURL string
}

func (s *LocalStorage) Enabled() bool { return false }

func (s *LocalStorage) Upload(ctx context.Context, filename string, body io.Reader, contentType string) (string, error) {
	_ = ctx
	_ = body
	_ = contentType
	return "/static/placeholder-logo.png", nil
}
