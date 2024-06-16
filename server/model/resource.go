package model

import (
	"fmt"

	"github.com/stsg/gophkeeper2/pkg/model"
	"github.com/stsg/gophkeeper2/pkg/model/enum"
)

type Resource struct {
	UserId int32  `db:"user_id"`
	Data   []byte `db:"data"`
	ResourceDescription
}

type ResourceDescription struct {
	Id   int32             `db:"id"`
	Meta []byte            `db:"meta"`
	Type enum.ResourceType `db:"type"`
}

func (rd *ResourceDescription) String() string {
	return fmt.Sprintf("[%d]: %v - %s", rd.Id, model.TypeToArg[rd.Type], string(rd.Meta))
}
