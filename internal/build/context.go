package build

import (
	"archive/tar"
	"bytes"
	"io"
	"time"
)

// BuildContext assembles a Docker build context as an in-memory tar archive
type BuildContext struct {
	buf *bytes.Buffer
	tw  *tar.Writer
}

func NewBuildContext() *BuildContext {
	buf := new(bytes.Buffer)
	return &BuildContext{
		buf: buf,
		tw:  tar.NewWriter(buf),
	}
}

// AddFile adds a file to the build context
func (bc *BuildContext) AddFile(name string, content []byte, mode int64) error {
	header := &tar.Header{
		Name:    name,
		Size:    int64(len(content)),
		Mode:    mode,
		ModTime: time.Now(),
	}
	if err := bc.tw.WriteHeader(header); err != nil {
		return err
	}
	_, err := bc.tw.Write(content)
	return err
}

// AddFileFromReader adds a file from a reader with known size
func (bc *BuildContext) AddFileFromReader(name string, r io.Reader, size int64, mode int64) error {
	header := &tar.Header{
		Name:    name,
		Size:    size,
		Mode:    mode,
		ModTime: time.Now(),
	}
	if err := bc.tw.WriteHeader(header); err != nil {
		return err
	}
	_, err := io.Copy(bc.tw, r)
	return err
}

// Reader finalizes the tar and returns a reader for the build context
func (bc *BuildContext) Reader() (io.Reader, error) {
	if err := bc.tw.Close(); err != nil {
		return nil, err
	}
	return bc.buf, nil
}
