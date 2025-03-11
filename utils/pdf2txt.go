package utils

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

func RunShellCMDPdf2Txt(log *log.Logger, fileLocation string, filename string) (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	extension := filepath.Ext(filename)
	fileNameTxt := filename[0:len(filename)-len(extension)] + ".txt"
	tmpPath := filepath.Join(dir, "tmp")
	tmpNewFilePath := filepath.Join(dir, "tmp", fileNameTxt)
	cmd := exec.Command("pdf2txt", "-out", tmpPath, fileLocation)
	err = cmd.Run()
	return tmpNewFilePath, err
}
