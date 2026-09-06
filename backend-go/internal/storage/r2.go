package storage

import (
	"bytes"
	"context"
	"errors"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// R2Store تخزين متوافق مع S3 (Cloudflare R2). ينشتغل بس إذا انضبطت
// متغيرات البيئة الأربعة، وإلا النظام يرجع للقرص المحلي.
type R2Store struct {
	client *s3.Client
	bucket string
}

// R2Config إعدادات الاتصال.
type R2Config struct {
	Bucket    string
	AccessKey string
	SecretKey string
	Endpoint  string
}

// Configured هل الإعدادات كاملة. ناقصة = ما نحاول نتصل أصلاً.
func (c R2Config) Configured() bool {
	return c.Bucket != "" && c.AccessKey != "" && c.SecretKey != "" && c.Endpoint != ""
}

func NewR2Store(cfg R2Config) (*R2Store, error) {
	if !cfg.Configured() {
		return nil, errors.New("إعدادات R2 ناقصة")
	}
	client := s3.New(s3.Options{
		// R2 ما يستخدم مناطق، بس الـSDK يطلب وحدة — auto هي المتعارف عليها
		Region:       "auto",
		BaseEndpoint: aws.String(cfg.Endpoint),
		Credentials:  credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, ""),
		// R2 ما يدعم أسلوب النطاق الفرعي للبكت
		UsePathStyle: true,
	})
	return &R2Store{client: client, bucket: cfg.Bucket}, nil
}

func (s *R2Store) Kind() string { return "R2 (bucket: " + s.bucket + ")" }

func (s *R2Store) Put(ctx context.Context, key string, data []byte, contentType string) error {
	ctx, cancel := ctxTimeout(ctx)
	defer cancel()
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
	})
	return err
}

func (s *R2Store) Get(ctx context.Context, key string) ([]byte, string, error) {
	ctx, cancel := ctxTimeout(ctx)
	defer cancel()
	out, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		var nsk *types.NoSuchKey
		if errors.As(err, &nsk) {
			return nil, "", ErrNotFound
		}
		return nil, "", err
	}
	defer out.Body.Close()

	data, err := readAllLimited(out.Body, MaxFileBytes)
	if err != nil {
		return nil, "", err
	}
	contentType := "application/octet-stream"
	if out.ContentType != nil && *out.ContentType != "" {
		contentType = *out.ContentType
	}
	return data, contentType, nil
}

func (s *R2Store) Delete(ctx context.Context, key string) error {
	ctx, cancel := ctxTimeout(ctx)
	defer cancel()
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	return err
}

var _ io.Reader = (*bytes.Reader)(nil)
