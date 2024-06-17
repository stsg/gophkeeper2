package configs

import (
	"os"
	"testing"
)

// Successfully reads and parses a valid configuration file
func TestInitAppConfig_SuccessfullyReadsAndParsesValidConfigFile(t *testing.T) {
	err := os.Chdir("../..")
	if err != nil {
		panic(err)
	}

	configContent := `{"ServerPort": ":42000", "PrivateKeyPath": "../../cert/client-key.pem"}`
	configFilePath := "test_config.json"
	err = os.WriteFile(configFilePath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}
	defer os.Remove(configFilePath)

	config, err := InitAppConfig(configFilePath)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if config.ServerPort != ":42000" {
		t.Errorf("Expected ServerPort to be ':42000', got '%s'", config.ServerPort)
	}
	if config.PrivateKeyPath != "cert/client-key.pem" {
		t.Errorf("Expected PrivateKeyPath to be 'cert/client-key.pem', got '%s'", config.PrivateKeyPath)
	}
}

// Configuration file path is empty
func TestInitAppConfig_EmptyConfigFilePath(t *testing.T) {
	config, err := InitAppConfig("")
	if config != nil {
		t.Errorf("Expected config to be nil, got %v", config)
	}
	if err == nil {
		t.Fatalf("Expected an error, got nil")
	}
	expectedError := "failed to init configuration: file path is not specified"
	if err.Error() != expectedError {
		t.Errorf("Expected error message '%s', got '%s'", expectedError, err.Error())
	}
}

func TestReadConfig(t *testing.T) {
	configContent := `{"ServerPort": ":42000", "PrivateKeyPath": "../../cert/client-key.pem"}`
	configFilePath := "test_config.json"
	err := os.WriteFile(configFilePath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}
	defer os.Remove(configFilePath)
}
