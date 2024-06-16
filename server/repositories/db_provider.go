package repositories

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/jackc/tern/migrate"
	"go.uber.org/zap"

    "github.com/stsg/gophkeeper2/pkg/logger"
	"github.com/stsg/gophkeeper2/server/configs"
	"github.com/stsg/gophkeeper2/server/model/errs"
)

const (
	migrationsDir = "./migrations/postgres/"
)

//go:generate mockgen -source=db_provider.go -destination=../mocks/repositories/db_provider.go -package=repositories

type DBProvider interface {
	HealthCheck(ctx context.Context) error
	GetConnection(ctx context.Context) (*pgxpool.Conn, error)
}

type pgProvider struct {
	log  *zap.SugaredLogger
	conn *pgxpool.Pool
}

func NewPgProvider(ctx context.Context, appConfig *configs.AppConfig) (DBProvider, error) {
	log := logger.NewLogger("pg-provider")
	if appConfig == nil {
		log.Error("Postgres DB config is empty")
		return nil, errors.New("failed to init pg repository: appConfig is nil")
	}
	pg := &pgProvider{log: log}
	err := pg.connect(ctx, appConfig.DBConnection, appConfig.DBMaxConnections)
	if err != nil {
		return nil, errs.DbError{Err: err}
	}
	err = pg.migrationUp(ctx)
	if err != nil {
		return nil, errs.DbError{Err: err}
	}
	return pg, nil
}

func (p *pgProvider) connect(ctx context.Context, connString string, maxConns int) error {
	if connString == "" {
		p.log.Error("Postgres DB config is empty")
		return errors.New("failed to init pg repository: dbConfig is empty")
	}
	p.log.Infof("Trying to connect: %s", connString)
	dbConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return errs.DbError{Err: err}
	}
	dbConfig.MaxConns = int32(maxConns)
	conn, err := pgxpool.ConnectConfig(ctx, dbConfig)
	if err != nil {
		return errs.DbError{Err: fmt.Errorf("failed to init pg repository: %v", err)}
	}
	p.conn = conn
	return nil
}

func (p *pgProvider) migrationUp(ctx context.Context) error {
	if p.conn == nil {
		return errs.InternalError{Err: errors.New("failed to start db migration: db connection is empty")}
	}
	acquireConn, err := p.conn.Acquire(ctx)
	if err != nil {
		return err
	}
	p.log.Info("Migrations are started")
	migrator, err := migrate.NewMigrator(ctx, acquireConn.Conn(), "schema_version")
	if err != nil {
		p.log.Errorf("Unable to create a migrator: %v\n", err)
		return errs.InternalError{Err: err}
	}
	err = migrator.LoadMigrations(migrationsDir)
	if err != nil {
		p.log.Errorf("Unable to load migrations: %v\n", err)
		return errs.InternalError{Err: err}
	}
	err = migrator.Migrate(ctx)
	if err != nil {
		p.log.Errorf("Unable to migrate: %v\n", err)
		return errs.InternalError{Err: err}
	}

	ver, err := migrator.GetCurrentVersion(ctx)
	if err != nil {
		return errs.DbError{Err: fmt.Errorf("failed to get current schema version: %v", err)}
	}
	p.log.Infof("Migration done. Current schema version: %d", ver)
	return nil
}

func (p *pgProvider) GetConnection(ctx context.Context) (*pgxpool.Conn, error) {
	acquireConn, err := p.conn.Acquire(ctx)
	if err != nil {
		return nil, errs.DbConnectionError{Err: err}
	}
	return acquireConn, err
}

func (p *pgProvider) HealthCheck(ctx context.Context) error {
	conn, err := p.GetConnection(ctx)
	if err != nil {
		p.log.Error("failed to check connection to Postgres DB: %v", err)
		return errs.InternalError{Err: err}
	}
	defer conn.Release()
	err = conn.Conn().Ping(ctx)
	if err != nil {
		p.log.Error("failed to check connection to Postgres DB: %v", err)
		return errs.InternalError{Err: err}
	}
	p.log.Info("Postgres DB connection is active")
	return nil
}
