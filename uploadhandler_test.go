package main

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUploadHandler(t *testing.T) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add a file to the request
	fileWriter, _ := writer.CreateFormFile("file", "example.txt")
	fileContents := []byte("Hello, world!")
	_, err := fileWriter.Write(fileContents)

	require.NoError(t, err)
	require.NoError(t, writer.Close())

	request := httptest.NewRequest(http.MethodPost, "/files", body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	uploadHandler, err := NewUploadHandler(request, "file", 10*1024*1024)
	require.NoError(t, err)

	err = uploadHandler.Handle(func(chunk []byte) error {
		return nil
	})
	require.NoError(t, err)
}
