package image

import (
	"io"
	"net/http"
	"os"
	"strings"
)

func validTemplatesUri(templatesUri string) bool {
	validPrefixes := []string{"http://", "https://", "file://"}
	for _, prefix := range validPrefixes {
		if strings.HasPrefix(templatesUri, prefix) {
			return true
		}
	}
	return false
}

// DownloadFile will download a url and store it in local filepath.
// It writes to the destination file as it downloads it, without
// loading the entire file into memory.
func downloadFile(url string, filepath string) error {
	// Get directory and create it if it doesn't exist
	dir := filepath[:strings.LastIndex(filepath, "/")]
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return err
		}
	}

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func copyLocalFile(orig string, destination string) error {
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

func CopyFile(src, dest string, isHTTP bool) error {
	if isHTTP {
		return downloadFile(src, dest)
	} else {
		return copyLocalFile(src, dest)
	}
}
