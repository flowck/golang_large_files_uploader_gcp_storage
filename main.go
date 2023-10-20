package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/flowck/large_files_uploader_gcp_storage/logs"
)

type Config struct {
	BucketName         string `envconfig:"GCP_BUCKET_NAME"`
	MaxFileSizeInBytes int64  `envconfig:"MAX_FILE_SIZE_IN_BYTES"`
}

func main() {
	logger := logs.New(true)

	config := &Config{}
	if err := envconfig.Process("", config); err != nil {
		logger.Fatal(err)
	}

	router := echo.New()
	done := make(chan os.Signal, 1)
	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
	storageUploader, err := NewStorageClient(ctx)
	if err != nil {
		panic(err)
	}

	router.Use(loggerMiddleware(logger))
	router.Use(middleware.Recover())

	router.POST("/files", func(c echo.Context) error {
		var uploader *UploadHandler
		uploader, err = NewUploadHandler(c.Request(), "file", config.MaxFileSizeInBytes)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
		}

		storageWriter := storageUploader.Upload(c.Request().Context(), config.BucketName, uploader.fileName)
		defer func() { _ = storageWriter.Close() }()

		err = uploader.Handle(func(chunk []byte) error {
			if os.Getenv("GCP_STORAGE_ENABLED") == "enabled" {
				if _, err = storageWriter.Write(chunk); err != nil {
					return err
				}
			}

			return nil
		})
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
		}

		return c.NoContent(http.StatusCreated)
	})

	router.GET("/files/:file_name", func(c echo.Context) error {
		fileName := c.Param("file_name")

		url, err := storageUploader.GetFileUrl(config.BucketName, fileName)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
		}

		return c.JSON(http.StatusOK, map[string]string{
			"url": url,
		})
	})

	s := http.Server{
		Addr:    ":8080",
		Handler: router,
		BaseContext: func(listener net.Listener) context.Context {
			return ctx
		},
	}
	go func() {
		log.Println("Server is running")
		log.Println(s.ListenAndServe())
	}()

	<-done
	tCtx, tCancel := context.WithTimeout(context.Background(), time.Second*1)
	defer tCancel()
	if err = s.Shutdown(tCtx); err != nil {
		log.Fatalf("unable to exit gracefully: %v", err)
	}

	log.Println("Exited gracefully")
}

func loggerMiddleware(logger *logs.Logger) echo.MiddlewareFunc {
	if logger.Level.String() == logs.DebugLevel.String() {
		return middleware.Logger()
	}

	return middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:    true,
		LogStatus: true,
		LogMethod: true,
		LogError:  true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			if v.Status >= 400 {
				logger.WithFields(logs.Fields{
					"method": v.Method,
					"URI":    v.URI,
					"status": v.Status,
					"error":  v.Error,
				}).Error("request processed with an error")

				return nil
			}

			logger.WithFields(logs.Fields{
				"method": v.Method,
				"URI":    v.URI,
				"status": v.Status,
			}).Info("request processed with success")
			return nil
		},
	})
}
