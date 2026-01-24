package pgx

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path"
	"regexp"
	"strings"

	"github.com/jamillosantos/migrations/v2"
)

var (
	ErrInvalidMigrationDirection = errors.New("invalid migration direction")

	migrationFileNameRegexp = regexp.MustCompile(`^(\d+)_(.*?)(\.(do|undo|down|up))?\.sql$`)
)

type source struct {
	fs     fs.ReadDirFS
	folder string

	repo     migrations.Repository
	dbGetter func() PgxDB
}

type migration struct {
	description string
	doFile      string
	undoFile    string
}

func (s *source) Load(ctx context.Context) (migrations.Repository, error) {
	entries, err := fs.ReadDir(s.fs, s.folder)
	if err != nil {
		return migrations.Repository{}, fmt.Errorf("failed listing migrations files: %w", err)
	}

	migrationSet := make(map[string]*migration)
	for _, entry := range entries {
		id, description, t := parseSQLFile(entry)
		if id == "" { // Does not match
			continue
		}

		migrationEntry := migrationSet[id]
		if migrationEntry == nil {
			migrationEntry = &migration{
				description: description,
			}
			migrationSet[id] = migrationEntry
		}

		switch t {
		case "", "up", "do":
			if migrationEntry.doFile != "" {
				// TODO: Improve this error
				return migrations.Repository{}, fmt.Errorf("migration %s already defined by %s", entry.Name(), migrationEntry.doFile)
			}
			migrationEntry.doFile = path.Join(s.folder, entry.Name())
		case "down", "undo":
			if migrationEntry.undoFile != "" {
				// TODO: Improve this error
				return migrations.Repository{}, fmt.Errorf("migration %s already defined by %s", entry.Name(), migrationEntry.doFile)
			}
			migrationEntry.undoFile = path.Join(s.folder, entry.Name())
		default:
			return migrations.Repository{}, fmt.Errorf("%w: %s (%s)", ErrInvalidMigrationDirection, t, entry.Name())
		}
	}

	for migrationID, migration := range migrationSet {
		m, err := s.repo.ByID(migrationID)
		if errors.Is(err, migrations.ErrMigrationNotFound) {
			// If the migration was not added yet, create the instance and add it.
			m = &migrationPgx{
				dbGetter: s.dbGetter,

				id:          migrationID,
				description: migration.description,
			}
		} else if err != nil {
			return migrations.Repository{}, err
		}

		mPgx := m.(*migrationPgx)

		mPgx.doFile = migration.doFile
		mPgx.doFileContent, err = loadMigrationFile(s.fs, migration.doFile)
		if err != nil {
			return migrations.Repository{}, err
		}
		mPgx.undoFile = migration.undoFile
		mPgx.undoFileContent, err = loadMigrationFile(s.fs, migration.undoFile)
		if err != nil {
			return migrations.Repository{}, err
		}
		err = s.repo.Add(m)
		if err != nil {
			return migrations.Repository{}, err
		}

		// Interrupt the loop if the context is done.
		select {
		case <-ctx.Done():
			return migrations.Repository{}, ctx.Err()
		default:
		}
	}

	return s.repo, nil
}

func (s *source) Add(_ context.Context, migration migrations.Migration) error {
	return s.repo.Add(migration)
}

func loadMigrationFile(fs fs.ReadDirFS, file string) (string, error) {
	if file == "" {
		// does not have migration
		return "", nil
	}

	f, err := fs.Open(file)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = f.Close()
	}()
	var buf strings.Builder
	_, err = io.Copy(&buf, f)
	if err != nil {
		return "", fmt.Errorf("cannot read migration file: %s: %w", file, err)
	}
	return buf.String(), nil
}

// parseSQLFile checks if the entry is valid and returns its id, description and type.
func parseSQLFile(entry fs.DirEntry) (id, description, t string) {
	if entry.IsDir() {
		return
	}
	m := migrationFileNameRegexp.FindStringSubmatch(entry.Name())
	if len(m) == 0 {
		return
	}
	id, description, t = m[1], strings.ReplaceAll(m[2], "_", " "), m[4]
	return
}
