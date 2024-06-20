package resources

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Format_Test(t *testing.T) {
	file := File{
		Name:      "example",
		Extension: "txt",
		Size:      1024,
	}

	description := "A sample text file"
	expected := "name: example\next: txt\nsize: 1024 bytes\ndescriptor: A sample text file\n"
	result := file.Format(description)

	assert.Equal(t, expected, result)
}
