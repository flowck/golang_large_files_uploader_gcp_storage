package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
)

type UploadHandler struct {
	fileReader io.Reader
	fileName   string
}

func NewUploadHandler(r *http.Request, formFieldName string) (*UploadHandler, error) {
	multipartReader, err := r.MultipartReader()
	if err != nil {
		return nil, err
	}

	var fileReader io.Reader
	var fileName string
	for {
		part, err := multipartReader.NextPart()
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
		fileReader: fileReader,
		fileName:   fileName,
	}, nil
}

func (h UploadHandler) Handle(handler func(chunk []byte) error) error {
	buffer := make([]byte, 4096)
	for {
		bytesRead, err := h.fileReader.Read(buffer)
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return fmt.Errorf("unable to read file: %v", err)
		}

		err = handler(buffer[:bytesRead])
		if err != nil {
			return fmt.Errorf("an error occurred while calling the handler: %v", err)
		}
	}

	return nil
}
