package resources

import (
	"fmt"

	"github.com/stsg/gophkeeper2/pkg/model/enum"
)

type LoginPassword struct {
	Login    string `json:"login,omitempty"`
	Password string `json:"password,omitempty"`
}

func NewLoginPassword(login string, password string) *LoginPassword {
	return &LoginPassword{Login: login, Password: password}
}

func (p *LoginPassword) Format(description string) string {
	return fmt.Sprintf("login: %s\npassword: %s\ndescription: %s", p.Login, p.Password, description)
}

func (p *LoginPassword) Type() enum.ResourceType {
	return enum.LoginPassword
}
