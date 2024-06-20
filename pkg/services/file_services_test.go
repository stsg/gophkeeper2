package services

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stsg/gophkeeper2/server/model/errs"
)

// Successfully creates a file at the specified path
func TestSaveFileSuccessfullyCreatesFile(t *testing.T) {
	path := "testfile.txt"
	chunks := make(chan []byte, 1)
	chunks <- []byte("test data")
	close(chunks)

	fm := NewFileService().(*fileService)
	errCh, err := fm.SaveFile(path, chunks)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	select {
	case err := <-errCh:
		t.Fatalf("expected no error, got %v", err)
	case <-time.After(1 * time.Second):
		// No error received, as expected
	}

	_, err = os.Stat(path)
	if os.IsNotExist(err) {
		t.Fatalf("expected file to be created, but it does not exist")
	}

	os.Remove(path)
}

// Error occurs when creating the file
func TestSaveFileErrorCreatingFile(t *testing.T) {
	path := "/invalidpath/testfile.txt"
	chunks := make(chan []byte, 1)
	chunks <- []byte("test data")
	close(chunks)

	fm := NewFileService().(*fileService)
	errCh, err := fm.SaveFile(path, chunks)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if errCh != nil {
		t.Fatalf("expected nil error channel, got %v", errCh)
	}
}

// Successfully reads a file and returns its content in chunks
func TestFileService_ReadFile_Success(t *testing.T) {
	// Create a temporary file with some content
	tempFile, err := os.CreateTemp("", "testfile")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	content := []byte("Hello, World!")
	if _, err := tempFile.Write(content); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	tempFile.Close()

	// Initialize the file service
	fs := NewFileService()

	// Create an error channel
	errCh := make(chan error)

	// Call the ReadFile method
	buf, stat, err := fs.ReadFile(tempFile.Name(), errCh)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify the file info
	if stat.Size() != int64(len(content)) {
		t.Errorf("expected file size %d, got %d", len(content), stat.Size())
	}

	// Read from the buffer channel and verify the content
	readContent := make([]byte, 0)
	for chunk := range buf {
		readContent = append(readContent, chunk...)
	}

	if string(readContent) != string(content) {
		t.Errorf("expected content %s, got %s", string(content), string(readContent))
	}
}

// Handles file open errors gracefully
func TestFileService_ReadFile_FileOpenError(t *testing.T) {
	// Initialize the file service
	fs := NewFileService()

	// Create an error channel
	errCh := make(chan error)

	// Call the ReadFile method with a non-existent file path
	_, _, err := fs.ReadFile("non_existent_file.txt", errCh)

	// Verify that an error is returned
	if err == nil {
		t.Fatalf("expected an error, got nil")
	}

	// Verify that the error is of type FileProcessingError
	if _, ok := err.(errs.FileProcessingError); !ok {
		t.Errorf("expected FileProcessingError, got %T", err)
	}
}

func TestNewFileServiceReturnsValidInstance(t *testing.T) {
	service := NewFileService()
	assert.NotNil(t, service, "Expected a valid FileService instance")
}
