package pgx

import (
	"context"
	"fmt"

	"github.com/jamillosantos/migrations/v2"
)

type migrationPgx struct {
	dbGetter func() PgxDB

	id              string
	description     string
	next            migrations.Migration
	previous        migrations.Migration
	doFile          string
	doFileContent   string
	undoFile        string
	undoFileContent string
}

// ID identifies the migration. Through the ID, all the sorting is done.
func (migration *migrationPgx) ID() string {
	return migration.id
}

// String will return a representation of the migration into a string format
// for user identification.
func (migration *migrationPgx) String() string {
	if migration.CanUndo() {
		return fmt.Sprintf("[%s,%s]", migration.doFile, migration.undoFile)
	}
	return fmt.Sprintf("[%s]", migration.doFile)
}

// Description is the humanized description for the migration.
func (migration *migrationPgx) Description() string {
	return migration.description
}

// Next will link this migration with the next. This link should be created
// by the source while it is being loaded.
func (migration *migrationPgx) Next() migrations.Migration {
	return migration.next
}

// SetNext will set the next migration
func (migration *migrationPgx) SetNext(value migrations.Migration) migrations.Migration {
	migration.next = value
	return migration
}

// Previous will link this migration with the previous. This link should be
// created by the Source while it is being loaded.
func (migration *migrationPgx) Previous() migrations.Migration {
	return migration.previous
}

// SetPrevious will set the previous migration
func (migration *migrationPgx) SetPrevious(value migrations.Migration) migrations.Migration {
	migration.previous = value
	return migration
}

func (migration *migrationPgx) executeSQL(ctx context.Context, sql string) error {
	db := migration.dbGetter()

	_, err := db.Exec(ctx, sql)
	if err != nil {
		return migrations.NewQueryError(err, sql)
	}
	return nil
}

// Do will execute the migration.
func (migration *migrationPgx) Do(ctx context.Context) error {
	return migration.executeSQL(ctx, migration.doFileContent)
}

// CanUndo is a flag that mark this flag as undoable.
func (migration *migrationPgx) CanUndo() bool {
	return migration.undoFile != ""
}

// Undo will undo the migration.
func (migration *migrationPgx) Undo(ctx context.Context) error {
	return migration.executeSQL(ctx, migration.undoFileContent)
}
