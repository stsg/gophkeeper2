package resources

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBankCard_Format(t *testing.T) {
	tests := []struct {
		name           string
		res            BankCard
		descr          string
		expectedString string
	}{
		{
			name: "Hate tests",
			res: BankCard{
				Number:   "123",
				ExpireAt: "12/34",
				Name:     "Name",
				Surname:  "Surname",
			},
			descr:          "descr",
			expectedString: "number: 123\nexpireAt: 12/34\nname: Name\nsurname: Surname\ndescription: descr",
		},
		{
			name: "Hate tests2",
			res: BankCard{
				Number:   "",
				ExpireAt: "",
				Name:     "",
				Surname:  "",
			},
			descr:          "",
			expectedString: "number: \nexpireAt: \nname: \nsurname: \ndescription: ",
		},
	}

	for _, test := range tests {
		fmt.Println(test.res.Format(test.descr))
		assert.Equal(t, test.expectedString, test.res.Format(test.descr))
	}
}

func TestFile_Format(t *testing.T) {
	tests := []struct {
		name           string
		res            File
		descr          string
		expectedString string
	}{
		{
			name: "Hate tests",
			res: File{
				Name:      "Name",
				Extension: ".boriest",
				Size:      123,
			},
			descr:          "descr",
			expectedString: "name: Name\next: .boriest\nsize: 123 bytes\ndescriptor: descr\n",
		},
		{
			name: "Hate tests2",
			res: File{
				Name:      "Name123",
				Extension: "wtf",
				Size:      -123,
			},
			descr:          "",
			expectedString: "name: Name123\next: wtf\nsize: -123 bytes\ndescriptor: \n",
		},
	}

	for _, test := range tests {
		fmt.Println(test.res.Format(test.descr))
		assert.Equal(t, test.expectedString, test.res.Format(test.descr))
	}
}

func TestLoginPassword_Format(t *testing.T) {
	tests := []struct {
		name           string
		res            LoginPassword
		descr          string
		expectedString string
	}{
		{
			name: "Hate tests",
			res: LoginPassword{
				Login:    "sesurity",
				Password: "qwerty",
			},
			descr:          "descr",
			expectedString: "login: sesurity\npassword: qwerty\ndescription: descr",
		},
		{
			name: "Hate tests2",
			res: LoginPassword{
				Login:    "again?",
				Password: "top secret!@#$%ABC123",
			},
			descr:          "",
			expectedString: "login: again?\npassword: top secret!@#$%ABC123\ndescription: ",
		},
	}

	for _, test := range tests {
		fmt.Println(test.res.Format(test.descr))
		assert.Equal(t, test.expectedString, test.res.Format(test.descr))
	}
}
