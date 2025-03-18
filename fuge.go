package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func main() {
	basePath, err := os.Executable()
	if err != nil {
		log.Printf("ERROR: can't get Exec path??? %s", err)
		basePath, err = os.Getwd()

		if err != nil {
			log.Fatalf("Failed to get executable path: %v", err)
		}
	} else {
		basePath = filepath.Dir(basePath)
	}

	pathcfg := filepath.Join(basePath, "config.ini")
	uploadDir := filepath.Join(basePath, "uploads")
	downloadDir := filepath.Join(basePath, "downloads")

	ensureDirectory(uploadDir)
	ensureDirectory(downloadDir)

	if len(os.Args) < 3 {
		fmt.Println("Usage:")
		fmt.Println("    fuge.exe download <exe-name>")
		fmt.Println("    fuge.exe upload <exe-name>")
		os.Exit(1)
	}

	awsConfig, err := loadAWSConfig(pathcfg)
	if err != nil {
		log.Fatalf("Failed to load AWS config file: %v", err)
	}

	action := os.Args[1]
	fileName := os.Args[2]

	objectKey := "addclan/" + fileName

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(awsConfig.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(awsConfig.AccessKey, awsConfig.SecretKey, "")),
	)

	if err != nil {
		log.Fatalf("Failed to load AWS config: %v", err)
	}

	s3Client := s3.NewFromConfig(cfg)

	switch action {
	case "upload":
		pathupl := filepath.Join(uploadDir, fileName)

		if err := uploadToS3(s3Client, awsConfig.Bucket, objectKey, pathupl); err != nil {
			log.Fatalf("Upload failed: %v", err)
		}
	case "download":
		pathdl := filepath.Join(downloadDir, fileName)

		if err := downloadFromS3(s3Client, awsConfig.Bucket, objectKey, pathdl); err != nil {
			log.Fatalf("Download failed: %v", err)
		}
	default:
		fmt.Println("Invalid action. Use 'upload' or 'download'")
		os.Exit(1)
	}
}

type AWSConfig struct {
	AccessKey string
	SecretKey string
	Bucket    string
	Region    string
}

func loadAWSConfig(filename string) (*AWSConfig, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	config := &AWSConfig{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key, value := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
		switch key {
		case "AWS_ACCESS_KEY":
			config.AccessKey = value
		case "AWS_SECRET":
			config.SecretKey = value
		case "AWS_BUCKET":
			config.Bucket = value
		case "AWS_REGION":
			config.Region = value
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return config, nil
}
func uploadToS3(client *s3.Client, bucketName string, objectKey string, fileloc string) error {
	fmt.Println("Uploading file to Fuge Services...")

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
	_, err = client.PutObject(context.TODO(), input)
	if err != nil {
		return fmt.Errorf("failed to upload: %w", err)
	}

	fmt.Println("Upload successful!")
	return nil
}

func downloadFromS3(client *s3.Client, bucketName string, objectKey string, fileloc string) error {
	fmt.Println("Downloading from S3...")

	// Prepare the GetObject input
	input := &s3.GetObjectInput{
		Bucket: &bucketName,
		Key:    &objectKey,
	}

	// Get the object from S3
	resp, err := client.GetObject(context.TODO(), input)
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

	fmt.Println("Download complete!")
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
