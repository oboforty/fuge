package ui

import (
	"fmt"
	"image"

	"github.com/aarzilli/nucular"
	"github.com/aarzilli/nucular/style"
	"github.com/oboforty/fuge/store"
)

var inputText string
var labelText string = "Write a file's name above and press `download`. To upload a file, type a file's name that you placed in the uploads/ folder."
var inp nucular.TextEditor
var uldlProgress int = 0

func InitializeUI() {
	inp = nucular.TextEditor{
		Flags:  nucular.EditField,
		Maxlen: 128,
	}

	wnd := nucular.NewMasterWindowSize(
		0,
		"Fuge",
		image.Point{400, 280}, updatefn,
	)
	wnd.SetStyle(style.FromTheme(style.DarkTheme, 1.0))
	wnd.Main()
}

func updatefn(w *nucular.Window) {
	w.Row(30).Dynamic(1)
	w.Label("File name:", "LC")

	w.Row(30).Dynamic(1)
	inp.Edit(w)

	// if w.TreePush(nucular.TreeTab, "Section Name", false) {
	// 	// Add UI elements inside the collapsible section
	// 	w.Label("Hello inside tree!", "LC")

	// 	// Always call TreePop to close the section properly
	// 	w.TreePop()
	// }

	// w.Row(30).Static(380)
	// w.Label(labelText, "LC")

	w.Row(32).Static(350)
	w.Progress(&uldlProgress, 100, false)

	w.Row(120).Static(350)
	w.LabelWrap(labelText)

	w.Row(30).Static(100, 100)
	if w.ButtonText("Download") {
		fileName := string(inp.Buffer)

		labelText = fmt.Sprintf("Downloading %s from Store...", fileName)

		inp.Delete(0, len(fileName))

		if fileName == "patch" {
			// reserved keyword, self patch
			go patch()
		} else {
			go download("addclan/"+fileName, fileName)
		}
	}

	if w.ButtonText("Upload") {
		fileName := string(inp.Buffer)

		labelText = fmt.Sprintf("Uploading file %s to Fuge Services...", fileName)

		inp.Delete(0, len(fileName))
		go upload("addclan/"+fileName, fileName)
	}
}

func uploadProgress(p int) {
	uldlProgress = p
}

func download(objectKey, fileName string) {
	uldlProgress = 0

	if err := store.DownloadFromS3(objectKey, fileName); err != nil {
		labelText = fmt.Sprintf("ERROR: Download failed: %v", err)
	} else {
		labelText = fmt.Sprintf("Download complete! %s", "uploads/"+fileName)
	}
}

func upload(objectKey, fileName string) {
	uldlProgress = 0

	if err := store.UploadToS3(objectKey, fileName, uploadProgress); err != nil {
		labelText = fmt.Sprintf("ERROR: Upload failed: %v", err)

	} else {
		labelText = fmt.Sprintf("Upload successful! %s", "downloads/"+fileName)
	}
}

func patch() {
	if err := store.Patch(); err != nil {
		labelText = fmt.Sprintf("ERROR: Patch failed: %v", err)
	} else {
		labelText = fmt.Sprintf("Patch OK at %s", "see in downloads/ folder")
	}
}
