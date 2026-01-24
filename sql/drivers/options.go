package drivers

import (
	"context"
)

type driverOpts struct {
	Ctx          context.Context
	DatabaseName string
	TableName    string
}

type Option func(*driverOpts)

func WithDatabaseName(name string) Option {
	return func(opts *driverOpts) {
		opts.DatabaseName = name
	}
}

func WithTableName(name string) Option {
	return func(opts *driverOpts) {
		opts.TableName = name
	}
}

func WithContext(ctx context.Context) Option {
	return func(opts *driverOpts) {
		opts.Ctx = ctx
	}
}
