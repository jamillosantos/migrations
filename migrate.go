package migrations

import (
	"context"
)

type ActionPLanner func(source Source, target Target) Planner

type migrateOptions struct {
	Runner        *Runner
	Reporter      RunnerReporter
	RunnerOptions []RunnerOption
	Planner       ActionPLanner
}

type MigrateOption func(*migrateOptions)

func Migrate(ctx context.Context, source Source, target Target, opts ...MigrateOption) (executionResponse ExecutionResponse, errResponse error) {
	options := migrateOptions{
		Runner:        nil,
		Reporter:      nil,
		RunnerOptions: nil,
		Planner:       MigratePlanner,
	}

	for _, opt := range opts {
		opt(&options)
	}

	if options.Runner == nil {
		options.Runner = NewRunner(source, target, options.RunnerOptions...)
	}

	runner := options.Runner

	unlocker, err := runner.target.Lock(ctx)
	if err != nil {
		return ExecutionResponse{}, err
	}
	defer func() {
		err := unlocker.Unlock(detach(ctx))
		if errResponse == nil { // If the originalReturned error isn't set, set it to the `Unlock` error, in case it fails.
			errResponse = err
		}
	}()

	err = runner.target.Create(ctx)
	if err != nil {
		return ExecutionResponse{}, err
	}

	plan, err := options.Planner(runner.source, runner.target).
		Plan(ctx)
	if err != nil {
		return ExecutionResponse{}, err
	}

	return runner.Execute(ctx, &ExecuteRequest{
		Plan: plan,
	})
}

// WithRunner sets the reporter to be used by the migration process.
func WithRunner(runner *Runner) MigrateOption {
	return func(options *migrateOptions) {
		options.Runner = runner
	}
}

// WithRunnerOptions sets the options to be used by the runner created. If a `WithRunner` is used, this is ignored.
func WithRunnerOptions(opts ...RunnerOption) MigrateOption {
	return func(options *migrateOptions) {
		options.RunnerOptions = opts
	}
}

// WithPlanner sets the planner to be used by the migration process. Default planner is MigratePlanner.
func WithPlanner(planner ActionPLanner) MigrateOption {
	return func(options *migrateOptions) {
		options.Planner = planner
	}
}
