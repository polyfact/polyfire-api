package utils

import (
	"context"
	"io"
	"log"
	"os"

	"cloud.google.com/go/storage"
)

type GCSUploader struct {
	cl         *storage.Client
	projectID  string
	bucketName string
}

func InitGCS() GCSUploader {
	client, err := storage.NewClient(context.Background())
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	uploader := GCSUploader{
		cl:         client,
		projectID:  os.Getenv("GCS_PROJECT_ID"),
		bucketName: os.Getenv("GCS_BUCKET_NAME"),
	}

	return uploader
}

func (c GCSUploader) UploadFile(file io.Reader, path string) (string, error) {
	ctx := context.Background()

	wc := c.cl.Bucket(c.bucketName).Object(path).NewWriter(ctx)
	if _, err := io.Copy(wc, file); err != nil {
		return "", err
	}
	if err := wc.Close(); err != nil {
		return "", err
	}

	return "gcs://" + c.bucketName + "/" + path, nil
}
