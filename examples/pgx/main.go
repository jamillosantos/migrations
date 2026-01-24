package main

import (
	"context"
	"embed"

	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"

	"github.com/jamillosantos/migrations/v2"
	migrationpgx "github.com/jamillosantos/migrations/v2/pgx"
	"github.com/jamillosantos/migrations/v2/reporters"
)

//go:embed migrations/*.sql
var migrationsFolder embed.FS

func main() {
	logger, err := zap.NewDevelopmentConfig().Build()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	ctx := context.Background()

	conn, err := pgx.Connect(ctx, "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable")
	if err != nil {
		panic(err)
	}
	defer conn.Close(ctx)

	s, err := migrationpgx.SourceFromFS(func() migrationpgx.PgxDB {
		return conn
	}, migrationsFolder, "migrations")
	if err != nil {
		panic(err)
	}

	t, err := migrationpgx.NewTarget(ctx, conn)
	if err != nil {
		panic(err)
	}

	_, err = migrations.Migrate(ctx, s, t, migrations.WithRunnerOptions(migrations.WithReporter(reporters.NewZapReporter(logger))))
	if err != nil {
		panic(err)
	}
}
