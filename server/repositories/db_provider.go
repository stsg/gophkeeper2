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

// NewPgProvider creates a new instance of the DBProvider interface using the provided
// context and appConfig. It initializes a logger and checks if the appConfig is nil. If it is,
// it logs an error and returns an error indicating that the appConfig is nil. It then creates a
// new pgProvider struct and calls the connect and migrationUp methods to establish a connection
// to the PostgreSQL database and apply any pending migrations. If either of these methods
// returns an error, it returns an error indicating a database error. Otherwise, it returns the
// newly created pgProvider instance.
//
// Parameters:
// - ctx: The context.Context object used for cancellation and timeouts.
// - appConfig: The *configs.AppConfig object containing the database connection configuration.
//
// Returns:
// - DBProvider: The newly created DBProvider instance.
// - error: An error indicating any issues encountered during the initialization process.
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

// connect establishes a connection to a PostgreSQL database using the provided connection string and maximum number of connections.
//
// Parameters:
// - ctx: The context.Context object used for cancellation and timeouts.
// - connString: The connection string for the PostgreSQL database.
// - maxConns: The maximum number of connections to the database.
//
// Returns:
// - error: An error indicating any issues encountered during the connection process.
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

// migrationUp performs the database migration up operation.
//
// It acquires a connection from the PostgreSQL database connection pool, creates a migrator, loads migrations,
// and performs the migration. It also retrieves the current schema version after the migration is complete.
//
// Parameters:
// - ctx: The context.Context object for the migration operation.
//
// Returns:
// - error: An error indicating any issues encountered during the migration process.
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

// GetConnection retrieves a connection from the database connection pool.
//
// Parameters:
// - ctx: The context.Context object used for cancellation and timeouts.
// Return type(s): (*pgxpool.Conn, error)
func (p *pgProvider) GetConnection(ctx context.Context) (*pgxpool.Conn, error) {
	acquireConn, err := p.conn.Acquire(ctx)
	if err != nil {
		return nil, errs.DbConnectionError{Err: err}
	}
	return acquireConn, err
}

// HealthCheck checks the health of the Postgres DB connection.
//
// It acquires a connection from the connection pool, pings the connection to check its availability,
// logs the result, and releases the connection.
//
// Parameters:
// - ctx: The context.Context object used for cancellation and timeouts.
//
// Return type(s):
// - error: Returns an error if there was a problem acquiring a connection, pinging the connection, or releasing the connection.
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
