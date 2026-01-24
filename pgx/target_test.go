package pgx

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/ory/dockertest/v3"

	"github.com/jamillosantos/migrations/v2"
)

var _ = Describe("Target", func() {
	var (
		db     *pgx.Conn
		target *Target
		ctx    context.Context
		pool   *dockertest.Pool
		res    *dockertest.Resource
	)

	BeforeEach(func() {
		ctx = context.Background()

		var err error
		pool, err = dockertest.NewPool("")
		Expect(err).ToNot(HaveOccurred())

		res, err = pool.Run("postgres", "15-alpine", []string{"POSTGRES_PASSWORD=secret", "POSTGRES_DB=testdb"})
		Expect(err).ToNot(HaveOccurred())

		port := res.GetPort("5432/tcp")
		connStr := fmt.Sprintf("postgres://postgres:secret@localhost:%s/testdb?sslmode=disable", port)

		err = pool.Retry(func() error {
			db, err = pgx.Connect(ctx, connStr)
			if err != nil {
				return err
			}
			return db.Ping(ctx)
		})
		Expect(err).ToNot(HaveOccurred())

		target, err = NewTarget(ctx, db,
			WithTableName("_tests_migrations"),
		)
		Expect(err).ToNot(HaveOccurred())

		Expect(target.Create(ctx)).To(Succeed())
	})

	AfterEach(func() {
		fmt.Println("Cleaning up...")
		if db != nil {
			Expect(db.Close(ctx)).To(Succeed())
		}
		if pool != nil {
			Expect(pool.Purge(res)).To(Succeed())
		}
	})

	Describe("NewTarget", func() {
		It("should create target with database name from query", func() {
			target, err := NewTarget(ctx, db)
			Expect(err).ToNot(HaveOccurred())
			Expect(target.databaseName).To(Equal("testdb"))
			Expect(target.tableName).To(Equal(DefaultMigrationsTableName))
		})

		It("should create target with provided options", func() {
			target, err := NewTarget(ctx, db, WithDatabaseName("manual_db"), WithTableName("custom_migrations"))
			Expect(err).ToNot(HaveOccurred())
			Expect(target.databaseName).To(Equal("manual_db"))
			Expect(target.tableName).To(Equal("custom_migrations"))
		})
	})

	It("should add a new migration", func() {
		Expect(target.Add(ctx, "1")).To(Succeed())
		Expect(target.FinishMigration(ctx, "1")).To(Succeed())

		done, err := target.Done(ctx)
		Expect(err).ToNot(HaveOccurred())
		Expect(done).To(Equal([]string{"1"}))
	})

	It("should fail with dirty migration", func() {
		Expect(target.Add(ctx, "1")).To(Succeed())

		_, err := target.Done(ctx)
		Expect(err).To(MatchError(migrations.ErrDirtyMigration))
	})

	It("should remove a migration", func() {
		Expect(target.Add(ctx, "1")).To(Succeed())
		Expect(target.FinishMigration(ctx, "1")).To(Succeed())

		Expect(target.Remove(ctx, "1")).To(Succeed())

		done, err := target.Done(ctx)
		Expect(err).ToNot(HaveOccurred())
		Expect(done).To(BeEmpty())
	})

	It("should fail to remove non-existent migration", func() {
		err := target.Remove(ctx, "999")
		Expect(err).To(MatchError(migrations.ErrMigrationNotFound))
	})

	It("should handle locking", func() {
		unlocker, err := target.Lock(ctx)
		Expect(err).ToNot(HaveOccurred())
		Expect(unlocker).ToNot(BeNil())

		// Second lock should block or fail if we used try_lock, but here it's advisory_lock which blocks.
		// For testing purposes, we just ensure we can unlock.
		Expect(unlocker.Unlock(ctx)).To(Succeed())
	})
})
