package utils

import (
	"io"
	"os"
	"strings"
)

func CopyFile(orig string, destination string) error {
	// Get directory and create it if it doesn't exist
	dir := destination[:strings.LastIndex(destination, "/")]
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return err
		}
	}

	// Open original file
	originalFile, err := os.Open(orig)
	if err != nil {
		return err
	}
	defer originalFile.Close()

	// Create the file
	newFile, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer newFile.Close()

	// Copy the file
	_, err = io.Copy(newFile, originalFile)
	if err != nil {
		return err
	}

	return nil
}
