package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func main() {

	// Set up Minio client
	minioEndpoint := "localhost:9000"
	minioAccessKey := "waxzuLkMytoKWpOQ"
	minioSecretKey := "uDgn6zuTDIeQv5d8X3dcqHwQ2nVzZmOt"
	minioUseSSL := false

	minioClient, err := minio.New(minioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(minioAccessKey, minioSecretKey, ""),
		Secure: minioUseSSL,
	})
	if err != nil {
		log.Fatalln(err)
	}

	// Set up HTTP server with Gin
	r := gin.Default()

	r.POST("/upload", func(c *gin.Context) {
		file, err := c.FormFile("file")
		if err != nil {
			c.String(http.StatusBadRequest, "Bad request")
			return
		}

		// Open the file
		fileHandle, err := file.Open()
		if err != nil {
			c.String(http.StatusInternalServerError, "Internal server error")
			return
		}
		defer fileHandle.Close()

		// Upload the file to Minio
		_, err = minioClient.PutObject(c.Request.Context(), "my-bucket", file.Filename, fileHandle, file.Size, minio.PutObjectOptions{})
		if err != nil {
			fmt.Printf("%+v", err)
			c.String(http.StatusInternalServerError, "Internal server error")
			return
		}

		c.String(http.StatusOK, "File uploaded successfully")
	})

	// GET /files/:filename - Get file from Minio
	r.GET("/files/:filename", func(c *gin.Context) {
		filename := c.Param("filename")
		fmt.Println("fi: ", filename)
		// Check if file exists in Minio
		_, err := minioClient.StatObject(context.Background(), "my-bucket", filename, minio.StatObjectOptions{})
		if err != nil {
			c.String(http.StatusNotFound, "File not found")
			return
		}

		// Download the file from Minio
		obj, err := minioClient.GetObject(c.Request.Context(), "my-bucket", filename, minio.GetObjectOptions{})
		if err != nil {
			c.String(http.StatusInternalServerError, "Internal server error")
			return
		}
		defer obj.Close()

		// Stream the file to the client
		c.Header("Content-Type", "application/octet-stream")
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
		_, err = io.Copy(c.Writer, obj)
		if err != nil {
			c.String(http.StatusInternalServerError, "Internal server error")
			return
		}
	})

	// PUT /files/:filename - Update file in Minio
	r.PUT("/files/:filename", func(c *gin.Context) {
		filename := c.Param("filename")

		// Check if file exists in Minio
		_, err := minioClient.StatObject(context.Background(), "my-bucket", filename, minio.StatObjectOptions{})
		if err != nil {
			c.String(http.StatusNotFound, "File not found")
			return
		}

		// Open the file
		file, err := c.FormFile("file")
		if err != nil {
			c.String(http.StatusBadRequest, "Bad request")
			return
		}

		fileHandle, err := file.Open()
		if err != nil {
			c.String(http.StatusInternalServerError, "Internal server error")
			return
		}
		defer fileHandle.Close()

		// Upload the updated file to Minio
		_, err = minioClient.PutObject(c.Request.Context(), "my-bucket", filename, fileHandle, file.Size, minio.PutObjectOptions{})
		if err != nil {
			c.String(http.StatusInternalServerError, "Internal server error")
			return
		}

		c.String(http.StatusOK, "File updated successfully")
	})

	// DELETE /files/:filename - Delete file from Minio
	r.DELETE("/files/:filename", func(c *gin.Context) {
		filename := c.Param("filename")

		// Check if file exists in Minio
		_, err := minioClient.StatObject(context.Background(), "my-bucket", filename, minio.StatObjectOptions{})
		if err != nil {
			c.String(http.StatusNotFound, "File not found")
			return
		}

		// Delete the file from Minio
		err = minioClient.RemoveObject(c.Request.Context(), "my-bucket", filename, minio.RemoveObjectOptions{})
		if err != nil {
			c.String(http.StatusInternalServerError, "Internal server error")
			return
		}

		c.String(http.StatusOK, "File deleted successfully")
	})

	// GET /files/:filename/url - Get URL for file with time expiry
	r.GET("/files/:filename/url", func(c *gin.Context) {
		filename := c.Param("filename")

		// Check if file exists in Minio
		_, err := minioClient.StatObject(context.Background(), "my-bucket", filename, minio.StatObjectOptions{})
		if err != nil {
			c.String(http.StatusNotFound, "File not found")
			return
		}

		// Generate a URL for the file with a time expiry of 1 hour
		url, err := minioClient.PresignedGetObject(context.Background(), "my-bucket", filename, time.Hour, nil)
		if err != nil {
			c.String(http.StatusInternalServerError, "Internal server error")
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"url": url.String(),
		})
	})

	err = r.Run(":8080")
	if err != nil {
		log.Fatalln(err)
	}
}
