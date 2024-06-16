package repositories

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"go.uber.org/zap"

	"github.com/stsg/gophkeeper2/pkg/logger"
	"github.com/stsg/gophkeeper2/server/model"
	"github.com/stsg/gophkeeper2/server/model/consts"
	"github.com/stsg/gophkeeper2/server/model/errs"
)

//go:generate mockgen -source=user_repository.go -destination=../mocks/repositories/user_repository.go -package=repositories

type UserRepository interface {
	CreateUser(context.Context, *model.User) (int32, error)
	GetUser(ctx context.Context, username string) (*model.User, error)
}

type userRepository struct {
	log *zap.SugaredLogger
	db  DBProvider
}

func NewUserRepository(db DBProvider) UserRepository {
	return &userRepository{log: logger.NewLogger("auth-repo"), db: db}
}

func (r *userRepository) CreateUser(ctx context.Context, user *model.User) (int32, error) {
	r.log.Infof("Creating user '%s'", user.Username)
	conn, err := r.db.GetConnection(ctx)
	if err != nil {
		r.log.Errorf("failed to get db connection: %v", err)
		return 0, errs.DbError{Err: err}
	}
	defer conn.Release()

	queryRow := conn.QueryRow(ctx, "insert into users (username, password) values ($1, $2) returning id", user.Username, user.Password)
	var userId int32
	err = queryRow.Scan(&userId)
	var pgError *pgconn.PgError
	if errors.As(err, &pgError) && pgError.Code == consts.UniqueViolation {
		r.log.Errorf("failed to save user '%s': already exist", user.Username)
		return 0, errs.ErrUserAlreadyExist
	}
	if err != nil {
		r.log.Errorf("failed to save user '%s': %v", user.Username, err)
		return 0, errs.DbError{Err: fmt.Errorf("failed to save user '%s': %v", user.Username, err)}
	}
	return userId, nil
}

func (r *userRepository) GetUser(ctx context.Context, username string) (*model.User, error) {
	conn, err := r.db.GetConnection(ctx)
	if err != nil {
		r.log.Errorf("failed to get db connection: %v", err)
		return nil, errs.DbError{Err: err}
	}
	defer conn.Release()
	user := &model.User{}
	queryRow := conn.QueryRow(ctx, "select id, password from users where username = $1", username)
	err = queryRow.Scan(&user.Id, &user.Password)
	if errors.Is(err, pgx.ErrNoRows) {
		r.log.Warnf("User '%s' not found", username)
		return nil, errs.ErrUserNotFound
	}
	if err != nil {
		r.log.Errorf("failed to scan user row '%s': %v", user.Username, err)
		return nil, errs.DbError{Err: fmt.Errorf("failed to scan user row '%s': %v", user.Username, err)}
	}

	return user, nil
}
