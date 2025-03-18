package store

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var uploadDir, downloadDir string
var s3Client *s3.Client
var bucketName string

func Init(basePath string) error {

	pathcfg := filepath.Join(basePath, "config.ini")
	uploadDir = filepath.Join(basePath, "uploads")
	downloadDir = filepath.Join(basePath, "downloads")

	ensureDirectory(uploadDir)
	ensureDirectory(downloadDir)

	awsConfig, err := loadAWSConfig(pathcfg)

	if err != nil {
		return err
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(awsConfig.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(awsConfig.AccessKey, awsConfig.SecretKey, "")),
	)

	if err != nil {
		return err
	}

	s3Client = s3.NewFromConfig(cfg)
	bucketName = awsConfig.Bucket

	return nil
}

func UploadToS3(objectKey string, fileName string) error {
	fileloc := filepath.Join(uploadDir, fileName)

	// Open the file specified by fileloc
	file, err := os.Open(fileloc)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", fileloc, err)
	}
	defer file.Close()

	// Prepare the S3 PutObject input with the file directly
	input := &s3.PutObjectInput{
		Bucket: &bucketName,
		Key:    &objectKey,
		Body:   file, // Directly pass the file as the Body
	}

	// Upload the file to S3
	_, err = s3Client.PutObject(context.TODO(), input)
	if err != nil {
		return fmt.Errorf("failed to upload: %w", err)
	}

	return nil
}

func DownloadFromS3(objectKey string, fileName string) error {
	fileloc := filepath.Join(downloadDir, fileName)

	// Prepare the GetObject input
	input := &s3.GetObjectInput{
		Bucket: &bucketName,
		Key:    &objectKey,
	}

	// Get the object from S3
	resp, err := s3Client.GetObject(context.TODO(), input)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	// Create the file where the object will be saved
	file, err := os.Create(fileloc)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", fileloc, err)
	}
	defer file.Close()

	// Copy the content from the S3 object to the local file
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}
	return nil
}

func ensureDirectory(dirPath string) error {
	// Check if the directory already exists
	info, err := os.Stat(dirPath)
	if err == nil && info.IsDir() {
		// Directory exists, no action needed
		return nil
	}

	// If the directory doesn't exist, try to create it
	if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dirPath, err)
	}

	fmt.Printf("Directory created: %s\n", dirPath)
	return nil
}
