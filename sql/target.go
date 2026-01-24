package sql

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"

	"github.com/jamillosantos/migrations/v2"
	"github.com/jamillosantos/migrations/v2/sql/drivers"
)

type DBExecer interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

type DB interface {
	DBExecer
	Driver() driver.Driver
	BeginTx(ctx context.Context, tx *sql.TxOptions) (*sql.Tx, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}

type Target struct {
	driver    drivers.Driver
	db        DB
	tableName string
}

func (target *Target) FinishMigration(ctx context.Context, id string) error {
	return target.driver.FinishMigration(ctx, id)
}

func (target *Target) StartMigration(ctx context.Context, id string) error {
	return target.driver.StartMigration(ctx, id)
}

func NewTarget(db DB, options ...TargetOption) (*Target, error) {
	opts := targetOpts{
		driver:    nil,
		tableName: drivers.DefaultMigrationsTableName,
	}
	for _, opt := range options {
		err := opt(&opts)
		if err != nil {
			return nil, err
		}
	}

	if opts.driver == nil {
		driverOptions := append([]drivers.Option{
			drivers.WithTableName(opts.tableName),
		}, opts.driverOptions...)
		d, err := drivers.DriverFromDB(db, driverOptions...)
		if err != nil {
			return nil, err
		}
		opts.driver = d
	}

	return &Target{
		driver:    opts.driver,
		db:        db,
		tableName: opts.tableName,
	}, nil
}

func (target *Target) Create(ctx context.Context) error {
	_, err := target.db.ExecContext(ctx, fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (id text PRIMARY KEY, dirty bool default true)", target.tableName))
	return err
}

func (target *Target) Destroy(ctx context.Context) error {
	_, err := target.db.ExecContext(ctx, fmt.Sprintf("DROP TABLE IF EXISTS %s", target.tableName))
	return err
}

func (target *Target) Current(ctx context.Context) (string, error) {
	list, err := target.Done(ctx)
	if err != nil {
		return "", err
	}

	if len(list) == 0 {
		return "", migrations.ErrNoCurrentMigration
	}

	return list[len(list)-1], nil
}

func (target *Target) Done(ctx context.Context) ([]string, error) {
	rs, err := target.db.QueryContext(ctx, fmt.Sprintf("SELECT id, dirty FROM %s ORDER BY id ASC", target.tableName))
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rs.Close()
	}()

	var (
		id    string
		dirty bool
	)
	result := make([]string, 0)
	for rs.Next() {
		err := rs.Scan(&id, &dirty)
		if err != nil {
			return nil, err
		}
		if dirty {
			return nil, migrations.WrapMigrationID(migrations.ErrDirtyMigration, id)
		}
		result = append(result, id)
	}
	return result, nil
}

func (target *Target) Add(ctx context.Context, id string) error {
	return target.driver.Add(ctx, id)
}

func (target *Target) Remove(ctx context.Context, id string) error {
	return target.driver.Remove(ctx, id)
}

func (target *Target) Lock(ctx context.Context) (migrations.Unlocker, error) {
	return target.driver.Lock(ctx)
}
