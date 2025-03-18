package cli

import (
	"fmt"
	"log"
	"os"

	"github.com/oboforty/fuge/store"
)

func ExecuteCLI() {
	action := os.Args[1]

	if action == "help" {
		Help()
		os.Exit(0)
	}

	fileName := os.Args[2]

	objectKey := "addclan/" + fileName

	switch action {
	case "upload":
		fmt.Println("Uploading file to Fuge Services...")

		if err := store.UploadToS3(objectKey, fileName); err != nil {
			log.Fatalf("Upload failed: %v", err)
		}

		fmt.Printf("Upload successful! %s", "downloads/"+fileName)
	case "download":
		fmt.Println("Downloading from Store...")

		if err := store.DownloadFromS3(objectKey, fileName); err != nil {
			log.Fatalf("Download failed: %v", err)
		}

		fmt.Printf("Download complete! %s", "uploads/"+fileName)
	default:
		fmt.Println("Invalid action. Use 'upload' or 'download'")
		os.Exit(1)
	}
}

func Help() {
	fmt.Println("Usage:")
	fmt.Println("    fuge.exe download <exe-name>")
	fmt.Println("    fuge.exe upload <exe-name>")
	os.Exit(1)
}
