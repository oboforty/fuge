package store

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

var uploadDir, downloadDir string
var s3Client *s3.Client
var bucketName string

// 5MB (Amazon S3 minimum part size for multipart upload)
const partSize = 5 * 1024 * 1024

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

func UploadToS3(objectKey, fileName string, updateProgress func(int)) error {
	fileloc := filepath.Join(uploadDir, fileName)

	file, err := os.Open(fileloc)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", fileloc, err)
	}
	defer file.Close()

	finfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info %s: %w", fileloc, err)
	}
	maxSize := float64(finfo.Size())
	uploadedSize := float64(0)

	// Initiate multipart upload
	createResp, err := s3Client.CreateMultipartUpload(context.TODO(), &s3.CreateMultipartUploadInput{
		Bucket: &bucketName,
		Key:    &objectKey,
	})
	if err != nil {
		return fmt.Errorf("failed to initiate multipart upload: %w", err)
	}

	var completedParts []types.CompletedPart
	buffer := make([]byte, partSize)
	partNumber := int32(1)

	for {
		bytesRead, err := file.Read(buffer)
		if err != nil && err != io.EOF {
			return fmt.Errorf("error reading file: %w", err)
		}
		if bytesRead == 0 {
			break
		}

		bytesRead64 := int64(bytesRead)
		uploadedSize += float64(bytesRead)
		currentPartNumber := partNumber
		progressSize := int(100 * (float64(uploadedSize) / maxSize))

		if updateProgress != nil {
			go updateProgress(progressSize)
		}
		log.Printf("Upload Progress size=%d part=%d", bytesRead64, currentPartNumber)

		// Upload part
		uploadResp, err := s3Client.UploadPart(context.TODO(), &s3.UploadPartInput{
			Bucket:        &bucketName,
			Key:           &objectKey,
			UploadId:      createResp.UploadId,
			PartNumber:    &currentPartNumber,
			Body:          io.NopCloser(bytes.NewReader(buffer[:bytesRead])),
			ContentLength: &bytesRead64,
		})

		if err != nil {
			log.Printf("S3 MultiPart Upload Error: %s", err.Error())

			// Abort upload on error
			s3Client.AbortMultipartUpload(context.TODO(), &s3.AbortMultipartUploadInput{
				Bucket:   &bucketName,
				Key:      &objectKey,
				UploadId: createResp.UploadId,
			})
			return fmt.Errorf("failed to upload part: %w", err)
		}

		completedParts = append(completedParts, types.CompletedPart{
			ETag:       uploadResp.ETag,
			PartNumber: &currentPartNumber,
		})

		partNumber++
	}

	sort.Slice(completedParts, func(i, j int) bool {
		return *completedParts[i].PartNumber < *completedParts[j].PartNumber
	})

	// Complete multipart upload
	_, err = s3Client.CompleteMultipartUpload(context.TODO(), &s3.CompleteMultipartUploadInput{
		Bucket:   &bucketName,
		Key:      &objectKey,
		UploadId: createResp.UploadId,
		MultipartUpload: &types.CompletedMultipartUpload{
			Parts: completedParts,
		},
	})
	if err != nil {
		log.Printf("S3 MultiPart Upload Complete Error: %s", err.Error())

		return fmt.Errorf("failed to complete multipart upload: %w", err)
	}

	return nil
}

func DownloadFromS3(objectKey, fileName string) error {
	fileloc := filepath.Join(downloadDir, fileName)

	// Prepare the GetObject input
	input := &s3.GetObjectInput{
		Bucket: &bucketName,
		Key:    &objectKey,
	}

	// Get the object from S3
	resp, err := s3Client.GetObject(context.TODO(), input)
	if err != nil {
		log.Printf("S3 Download Error: %s", err.Error())

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
