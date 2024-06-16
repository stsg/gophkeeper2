package resources

import (
	"fmt"

	"github.com/stsg/gophkeeper2/pkg/model/enum"
)

type File struct {
	Name      string
	Extension string
	Size      int64
}

func (p *File) Format(description string) string {
	return fmt.Sprintf("name: %s\next: %s\nsize: %d bytes\ndescriptor: %s\n", p.Name, p.Extension, p.Size, description)
}

func (p *File) Type() enum.ResourceType {
	return enum.File
}
