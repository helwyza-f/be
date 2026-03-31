package storage

import (
	"context"
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

type S3Client struct {
	client *s3.Client
	bucket string
	region string
}

func NewS3Client() (*S3Client, error) {
	region := os.Getenv("AWS_REGION")
	bucket := os.Getenv("AWS_BUCKET")

	// Load config otomatis dari env AWS_ACCESS_KEY_ID & AWS_SECRET_ACCESS_KEY
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config, %v", err)
	}

	return &S3Client{
		client: s3.NewFromConfig(cfg),
		bucket: bucket,
		region: region,
	}, nil
}

func (s *S3Client) UploadFile(ctx context.Context, file *multipart.FileHeader, folder string) (string, error) {
	f, err := file.Open()
	if err != nil {
		return "", err
	}
	defer f.Close()

	// Rename file: folder/uuid-namaasli.ext
	// Contoh: tenants/minibos/a1b2c3d4-logo.png
	fileExt := filepath.Ext(file.Filename)
	fileName := strings.TrimSuffix(file.Filename, fileExt)
	newFileName := fmt.Sprintf("%s/%s-%s%s", folder, uuid.New().String(), fileName, fileExt)

	_, err = s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(newFileName),
		Body:        f,
		ContentType: aws.String(file.Header.Get("Content-Type")),
	})

	if err != nil {
		return "", fmt.Errorf("failed to upload object, %v", err)
	}

	// URL Public
	url := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.bucket, s.region, newFileName)
	return url, nil
}