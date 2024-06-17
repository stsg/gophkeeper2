package services

import (
	"context"
	"testing"

	"github.com/stsg/gophkeeper2/pkg/logger"
	"github.com/stsg/gophkeeper2/server/configs"
	"github.com/stsg/gophkeeper2/server/model"
	"github.com/stsg/gophkeeper2/server/repositories"
	"golang.org/x/crypto/bcrypt"
)

func TestValidatePassword_ValidPassword(t *testing.T) {
	// Initialize the logger
	log := logger.NewLogger("user-srv")

	// Create a user with a hashed password
	password := "validpassword"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 8)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}
	user := &model.User{
		Username: "testuser",
		Password: hashedPassword,
	}

	// Create a userService instance
	ctx, _ := context.WithCancel(context.Background())
	appConfig, _ := configs.InitAppConfig("cfg/config.json")
	dbProvider, _ := repositories.NewPgProvider(ctx, appConfig)
	repo := repositories.NewUserRepository(dbProvider) // Assuming a constructor for UserRepository
	service := &userService{log: log, repo: repo}

	// Validate the password
	isValid, err := service.ValidatePassword(context.Background(), user, password)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !isValid {
		t.Errorf("expected password to be valid")
	}
}

// Invalid password returns false and no error
func TestValidatePassword_InvalidPassword(t *testing.T) {
	// Initialize the logger
	log := logger.NewLogger("user-srv")

	// Create a user with a hashed password
	password := "validpassword"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 8)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}
	user := &model.User{
		Username: "testuser",
		Password: hashedPassword,
	}

	// Create a userService instance
	ctx, _ := context.WithCancel(context.Background())
	appConfig, _ := configs.InitAppConfig("cfg/config.json")
	dbProvider, _ := repositories.NewPgProvider(ctx, appConfig)
	repo := repositories.NewUserRepository(dbProvider) // Assuming a constructor for UserRepository
	service := &userService{log: log, repo: repo}

	// Validate the password with an invalid password
	isValid, err := service.ValidatePassword(context.Background(), user, "invalidpassword")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if isValid {
		t.Errorf("expected password to be invalid")
	}
}
