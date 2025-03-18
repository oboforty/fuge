package store

import (
	"bufio"
	"os"
	"strings"
)

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
