package main

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
)

var (
	ErrMaxSizeInBytesExceeded = errors.New("maxSizeInBytes has been exceeded")
)

type UploadHandler struct {
	fileReader     io.Reader
	fileName       string
	maxSizeInBytes int64
}

func NewUploadHandler(r *http.Request, formFieldName string, maxSizeInBytes int64) (*UploadHandler, error) {
	multipartReader, err := r.MultipartReader()
	if err != nil {
		return nil, err
	}

	var fileReader io.Reader
	var fileName string
	var part *multipart.Part
	for {
		part, err = multipartReader.NextPart()
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return nil, err
		}

		if part.FormName() == formFieldName {
			fileReader = part
			fileName = part.FileName()
			break
		}
	}

	if fileReader == nil {
		return nil, errors.New("no file attached to the request")
	}

	return &UploadHandler{
		fileReader:     fileReader,
		fileName:       fileName,
		maxSizeInBytes: maxSizeInBytes,
	}, nil
}

func (h UploadHandler) Handle(handler func(chunk []byte) error) error {
	var totalBytesRead int64 = 0
	buffer := make([]byte, 4096)

	for {
		bytesRead, err := h.fileReader.Read(buffer)
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return fmt.Errorf("unable to read file: %v", err)
		}

		if totalBytesRead > h.maxSizeInBytes {
			return ErrMaxSizeInBytesExceeded
		}

		err = handler(buffer[:bytesRead])
		if err != nil {
			return fmt.Errorf("an error occurred while calling the handler: %v", err)
		}

		totalBytesRead += int64(bytesRead)
	}

	return nil
}
