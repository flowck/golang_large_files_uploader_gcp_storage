package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"cloud.google.com/go/storage"
	"github.com/oklog/ulid/v2"
)

type StorageClient struct {
	client *storage.Client
}

type StorageWriter struct {
	writer *storage.Writer
}

func (w *StorageWriter) Close() error {
	return w.writer.Close()
}

func (w *StorageWriter) Write(data []byte) (int, error) {
	return w.writer.Write(data)
}

func NewStorageClient(ctx context.Context) (*StorageClient, error) {
	client, err := storage.NewClient(ctx, storage.WithJSONReads())
	if err != nil {
		return nil, err
	}

	return &StorageClient{
		client: client,
	}, nil
}

func (u *StorageClient) Upload(ctx context.Context, bucketName, fileName string) *StorageWriter {
	bucket := u.client.Bucket(bucketName)
	object := bucket.Object(fmt.Sprintf("%s-%s", ulid.Make().String(), fileName))
	writer := object.NewWriter(ctx)
	return &StorageWriter{writer: writer}
}

func (u *StorageClient) GetFileUrl(bucketName, fileName string) (string, error) {
	return u.client.Bucket(bucketName).SignedURL(fileName, &storage.SignedURLOptions{
		Method:  http.MethodGet,
		Expires: time.Now().UTC().Add(time.Hour * 24),
	})
}
