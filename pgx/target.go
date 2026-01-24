package pgx

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jamillosantos/migrations/v2"
	"github.com/spaolacci/murmur3"
)

const DefaultMigrationsTableName = "_migrations"

type Target struct {
	db           PgxDB
	tableName    string
	databaseName string
}

type targetOpts struct {
	tableName    string
	databaseName string
}

type TargetOption func(*targetOpts)

func WithTableName(tableName string) TargetOption {
	return func(opts *targetOpts) {
		opts.tableName = tableName
	}
}

func WithDatabaseName(databaseName string) TargetOption {
	return func(opts *targetOpts) {
		opts.databaseName = databaseName
	}
}

func NewTarget(ctx context.Context, db PgxDB, options ...TargetOption) (*Target, error) {
	opts := targetOpts{
		tableName: DefaultMigrationsTableName,
	}
	for _, opt := range options {
		opt(&opts)
	}

	if opts.databaseName == "" {
		rows, err := db.Query(ctx, "SELECT current_database()")
		if err != nil {
			return nil, fmt.Errorf("error obtaining current database name: %w", err)
		}
		defer rows.Close()

		if rows.Next() {
			err = rows.Scan(&opts.databaseName)
			if err != nil {
				return nil, fmt.Errorf("error scanning current database name: %w", err)
			}
		} else {
			return nil, fmt.Errorf("missing database name")
		}
	}

	return &Target{
		db:           db,
		tableName:    opts.tableName,
		databaseName: opts.databaseName,
	}, nil
}

func (t *Target) Current(ctx context.Context) (string, error) {
	list, err := t.Done(ctx)
	if err != nil {
		return "", err
	}

	if len(list) == 0 {
		return "", migrations.ErrNoCurrentMigration
	}

	return list[len(list)-1], nil
}

func (t *Target) Create(ctx context.Context) error {
	_, err := t.db.Exec(ctx, "CREATE TABLE IF NOT EXISTS "+t.tableName+" (id text PRIMARY KEY, dirty bool default true)")
	return err
}

func (t *Target) Destroy(ctx context.Context) error {
	_, err := t.db.Exec(ctx, "DROP TABLE IF EXISTS "+t.tableName)
	return err
}

func (t *Target) Done(ctx context.Context) ([]string, error) {
	rs, err := t.db.Query(ctx, "SELECT id, dirty FROM "+t.tableName+" ORDER BY id ASC")
	if err != nil {
		return nil, err
	}
	defer rs.Close()

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

func (t *Target) Add(ctx context.Context, id string) error {
	_, err := t.db.Exec(ctx, "INSERT INTO "+t.tableName+" (id, dirty) VALUES ($1, true)", id)
	if err != nil {
		return fmt.Errorf("failed adding migration to the executed list: %w", err)
	}
	return nil
}

func (t *Target) Remove(ctx context.Context, id string) error {
	tag, err := t.db.Exec(ctx, "DELETE FROM "+t.tableName+" WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed removing migration from the executed list: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return migrations.ErrMigrationNotFound
	}
	return nil
}

func (t *Target) FinishMigration(ctx context.Context, id string) error {
	tag, err := t.db.Exec(ctx, "UPDATE "+t.tableName+" SET dirty = false WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed finishing migration: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return migrations.ErrMigrationNotFound
	}
	return nil
}

func (t *Target) StartMigration(ctx context.Context, id string) error {
	tag, err := t.db.Exec(ctx, "UPDATE "+t.tableName+" SET dirty = true WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed starting migration: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return migrations.ErrMigrationNotFound
	}
	return nil
}

func (t *Target) Lock(ctx context.Context) (migrations.Unlocker, error) {
	tx, err := t.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed starting transaction for locking: %w", err)
	}

	advisoryLockID, err := t.generateLockID()
	if err != nil {
		_ = tx.Rollback(ctx)
		return nil, fmt.Errorf("failed locking database: %w", err)
	}
	_, err = tx.Exec(ctx, "SELECT pg_advisory_lock($1)", advisoryLockID)
	if err != nil {
		_ = tx.Rollback(ctx)
		return nil, fmt.Errorf("failed locking database: %w", err)
	}
	return &unlocker{tx: tx, lockID: advisoryLockID}, nil
}

func (t *Target) generateLockID() (int64, error) {
	h := murmur3.New64()
	if _, err := h.Write([]byte(t.databaseName)); err != nil {
		return 0, err
	}
	if _, err := h.Write([]byte("|||")); err != nil {
		return 0, err
	}
	if _, err := h.Write([]byte(t.tableName)); err != nil {
		return 0, err
	}
	return int64(h.Sum64()), nil
}

type unlocker struct {
	tx     pgx.Tx
	lockID int64
}

func (u *unlocker) Unlock(ctx context.Context) error {
	_, err := u.tx.Exec(ctx, "SELECT pg_advisory_unlock($1)", u.lockID)
	if err != nil {
		_ = u.tx.Rollback(ctx)
		return fmt.Errorf("failed unlocking migration: %w", err)
	}
	_ = u.tx.Commit(ctx)
	return nil
}
