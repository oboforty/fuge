package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/oboforty/fuge/cli"
	"github.com/oboforty/fuge/store"
	"github.com/oboforty/fuge/ui"
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

	if err = store.Init(basePath); err != nil {
		log.Fatalf("Failed to load AWS config file: %v", err)
	}

	if len(os.Args) < 2 {
		// UI mode
		ui.InitializeUI()
	} else {
		// cmd mode
		cli.ExecuteCLI()
	}
}
