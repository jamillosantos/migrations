package migrations

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrNonUniqueMigrationID       = errors.New("migration id is not unique")
	ErrMigrationNotFound          = errors.New("migration not found")
	ErrUnkownMigrations           = errors.New("unknown migrations were found")
	ErrNoCurrentMigration         = errors.New("no current migration")
	ErrCurrentMigrationNotFound   = errors.New("current migration not found in the list")
	ErrCurrentMigrationMoreRecent = errors.New("current migration is more recent than target migration")
	ErrNoMigrationsAvailable      = errors.New("no migrations available")
	ErrMigrationNotUndoable       = errors.New("migration cannot be undone")

	// ErrStepOutOfIndex is returned when a `StepResolver` cannot resolve a
	// migration due to the resolved index be outside of the migration list.
	ErrStepOutOfIndex = errors.New("step out of bounds")

	// ErrMigrationNotListed is returned when a migration is not found in the
	// `Source` list.
	ErrMigrationNotListed = errors.New("migration not in the source list")

	// ErrInvalidAction is returned when, while executing, the `Action.Action`
	// has an invalid value.
	ErrInvalidAction = errors.New("undefined action")

	// ErrInvalidPatternForFile is returned when a .sql does not meet the
	// nameing convention.
	ErrInvalidPatternForFile = errors.New("file does not match file name rule")

	ErrInvalidSQLFileNameDuplicated = errors.New("file name is duplicated")

	// ErrWrongExecutionContext is returned when an migration execution context
	// does not fit the configuration to be passed to a specific migration
	// implementation. For example, a SQL migration receives a executionContext
	// of a MongoDB.
	ErrWrongExecutionContext = errors.New("wrong execution context")
)

// MigrationCodeError wraps an error with a migration ID.
type MigrationIDError interface {
	error
	MigrationID() time.Time
	Unwrap() error
}

// MigrationError wraps an error with a migration property.
type MigrationError interface {
	error
	Migration() Migration
	Unwrap() error
}

// MigrationsError wraps an error with a list of migrations.
type MigrationsError interface {
	error
	Migrations() []Migration
}

type migrationIDError struct {
	error
	migrationID time.Time
}

type migrationError struct {
	error
	migration Migration
}

type migrationsError struct {
	error
	migrations []Migration
}

// WrapMigrationID creates a `MigrationCodeError` based on an existing error.
func WrapMigrationID(err error, migrationID time.Time) MigrationIDError {
	return &migrationIDError{
		err,
		migrationID,
	}
}

// WrapMigration creates a `MigrationError` based on an existing error.
func WrapMigration(err error, migration Migration) MigrationError {
	return &migrationError{
		err,
		migration,
	}
}

func WrapMigrations(err error, migrations ...Migration) MigrationsError {
	return &migrationsError{
		err,
		migrations,
	}
}

func (err *migrationIDError) MigrationID() time.Time {
	return err.migrationID
}

func (err *migrationIDError) Unwrap() error {
	return err.error
}

func (err *migrationIDError) Error() string {
	return "migration " + err.migrationID.Format(DefaultMigrationIDFormat) + ": " + err.error.Error()
}

func (err *migrationError) Migration() Migration {
	return err.migration
}

func (err *migrationError) Unwrap() error {
	return err.error
}

func (err *migrationError) Error() string {
	return err.migration.ID().Format(DefaultMigrationIDFormat) + ": " + err.error.Error()
}

func (err *migrationsError) Migrations() []Migration {
	return err.migrations
}

func (err *migrationsError) Error() string {
	var r strings.Builder
	for i, migration := range err.migrations {
		if i > 0 {
			r.WriteString(",")
		}
		r.WriteString(migration.ID().Format(DefaultMigrationIDFormat))
	}
	r.WriteString(": ")
	r.WriteString(err.error.Error())
	return r.String()
}

var (
	MigrationIDNone = time.Unix(0, 0)
)

type QueryError interface {
	Query() string
}

type queryError struct {
	error
	query string
}

func NewQueryError(err error, query string) error {
	return &queryError{err, query}
}

func (err *queryError) Unwrap() error {
	return err.error
}

func (err queryError) Query() string {
	return err.query
}
