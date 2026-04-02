# migrations

[![Go Reference](https://pkg.go.dev/badge/github.com/jamillosantos/migrations/v2.svg)](https://pkg.go.dev/github.com/jamillosantos/migrations/v2)

A flexible, driver-agnostic migration library for Go. It separates **what** migrations are (Source), **where** state is tracked (Target), and **how** they run (Runner) — so you can migrate SQL databases, pgx connections, or anything else with the same API.

## Features

- **Multiple migration sources** — SQL files via `embed.FS`, Go functions, or both combined
- **Bidirectional** — forward (`Do`) and backward (`Undo`) with optional undo support
- **Database drivers** — `database/sql` (PostgreSQL, SQLite) and native `pgx/v5`
- **Advisory locking** — PostgreSQL advisory locks prevent concurrent migration runs
- **Dirty state detection** — flags incomplete migrations so you know when something failed mid-run
- **Flexible planners** — migrate, reset, rewind, step forward/backward, do one, undo one
- **Progress reporting** — plug in Zap or your own reporter

## Install

```bash
go get github.com/jamillosantos/migrations/v2
```

## Quick Start

### SQL files with `database/sql`

```go
package main

import (
	"context"
	"database/sql"
	"embed"

	_ "github.com/lib/pq"

	"github.com/jamillosantos/migrations/v2"
	migrationsql "github.com/jamillosantos/migrations/v2/sql"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func main() {
	db, err := sql.Open("postgres", "postgres://localhost:5432/mydb?sslmode=disable")
	if err != nil {
		panic(err)
	}

	source, err := migrationsql.SourceFromFS(func() migrationsql.DBExecer {
		return db
	}, migrationsFS, "migrations")
	if err != nil {
		panic(err)
	}

	target, err := migrationsql.NewTarget(db)
	if err != nil {
		panic(err)
	}

	_, err = migrations.Migrate(context.Background(), source, target)
	if err != nil {
		panic(err)
	}
}
```

### SQL files with pgx

```go
package main

import (
	"context"
	"embed"

	"github.com/jackc/pgx/v5"

	"github.com/jamillosantos/migrations/v2"
	migrationpgx "github.com/jamillosantos/migrations/v2/pgx"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func main() {
	ctx := context.Background()

	conn, err := pgx.Connect(ctx, "postgres://localhost:5432/mydb?sslmode=disable")
	if err != nil {
		panic(err)
	}
	defer conn.Close(ctx)

	source, err := migrationpgx.SourceFromFS(func() migrationpgx.PgxDB {
		return conn
	}, migrationsFS, "migrations")
	if err != nil {
		panic(err)
	}

	target, err := migrationpgx.NewTarget(ctx, conn)
	if err != nil {
		panic(err)
	}

	_, err = migrations.Migrate(ctx, source, target)
	if err != nil {
		panic(err)
	}
}
```

### Go function migrations

For migrations that need to do more than run SQL (call APIs, transform data, etc.), use the `fnc` package:

```go
package mymigrations

import (
	"context"

	"github.com/jamillosantos/migrations/v2"
	"github.com/jamillosantos/migrations/v2/fnc"
)

var CodeMigrations []migrations.Migration

func register(do func(ctx context.Context) error) {
	CodeMigrations = append(CodeMigrations, fnc.Migration(do, fnc.WithSkip(2)))
}
```

Function-based and SQL-based migrations can be combined into a single source — they'll be sorted by timestamp and run in order.

## Migration File Format

SQL migration files follow this naming convention:

```
<timestamp>_<description>.<direction>.sql
```

Supported direction suffixes: `.do.sql`, `.up.sql`, `.undo.sql`, `.down.sql`

Examples:
```
20250315143000_create_users.do.sql
20250315143000_create_users.undo.sql
20250315144500_add_email_index.do.sql
```

Undo files are optional. Migrations without an undo file are forward-only.

## Generating Migration Files

Add the CLI as a tool dependency:

```bash
go get -tool github.com/jamillosantos/migrations/v2/cli/migrations
```

Then generate migration files:

```bash
go tool migrations create -destination=migrations
```

## Planners

Control how migrations run by passing a planner to `Migrate`:

| Planner | Behavior |
|---|---|
| `MigratePlanner` (default) | Run all pending migrations |
| `ResetPlanner` | Undo everything, then migrate to latest |
| `RewindPlanner` | Undo all applied migrations |
| `StepPlanner(n)` | Step forward (`n > 0`) or backward (`n < 0`) by N |
| `DoPlanner` | Run the next pending migration |
| `UndoPlanner` | Undo the last applied migration |

```go
migrations.Migrate(ctx, source, target,
	migrations.WithPlanner(migrations.ResetPlanner),
)
```

## Reporters

Track migration progress with a reporter:

```go
import "github.com/jamillosantos/migrations/v2/reporters"

migrations.Migrate(ctx, source, target,
	migrations.WithRunnerOptions(
		migrations.WithReporter(reporters.NewZapReporter(logger)),
	),
)
```

Built-in reporters: `ZapReporter`, `NoopReporter`. Implement `RunnerReporter` for custom reporting.

## Architecture

```
Source (what to run)          Target (state tracking)
  - SQL files (embed.FS)       - _migrations table
  - Go functions (fnc)         - Any custom store
  - Combined sources
         \                    /
          \                  /
           Runner + Planner
           (how to execute)
```

**Source** loads available migrations from whatever media stores them — embedded SQL files, Go code, or both.

**Target** tracks which migrations have been applied. For SQL databases, this is a `_migrations` table. You can implement `Target` for any storage backend.

**Runner** ties Source and Target together. A **Planner** decides which actions to take (migrate, reset, rewind, step), and the Runner executes them.

## License

MIT
