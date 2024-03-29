package migrations

import (
	"github.com/pkg/errors"
)

type migratePlanner struct {
	source Source
	target Target
}

func MigratePlanner(source Source, target Target) Planner {
	return &migratePlanner{
		source: source,
		target: target,
	}
}

func (planner *migratePlanner) Plan() (Plan, error) {
	list, err := planner.source.List()
	if err != nil {
		return nil, errors.Wrap(err, "error listing available migrations")
	}

	current, err := planner.target.Current()
	if errors.Is(err, ErrNoCurrentMigration) {
		plan := make(Plan, len(list))
		for i, m := range list {
			plan[i] = &Action{
				Action:    ActionTypeDo,
				Migration: m,
			}
		}
		// If there is no current migration, all migrations should run.
		return plan, nil
	} else if err != nil {
		// Otherwise, it is just an error.
		return nil, err
	}

	listM := make(map[string]Migration, len(list))
	for _, m := range list {
		listM[m.ID()] = m
	}

	done, err := planner.target.Done()
	if err != nil {
		return nil, errors.Wrap(err, "failed listing migrations applied")
	}
	for _, m := range done {
		if _, ok := listM[m.ID()]; !ok {
			return nil, errors.Wrap(ErrMigrationNotListed, m.String())
		}
	}

	for i, m := range done {
		if m.ID() != list[i].ID() {
			return nil, errors.Wrap(ErrStaleMigrationDetected, list[i].String())
		}
	}

	// This is the migration that we are trying to reach. Always the most recent one.
	targetMigration := list[len(list)-1]

	// If the current migration is the same as the target migration
	if current.ID() == targetMigration.ID() {
		// Nothing should be done.
		return Plan{}, nil
	}

	currentMigrationIndex, err := findMigrationIndex(list, current)
	if err != nil {
		return nil, err
	}

	// If the current migration is further in the future than the target migration.
	if current.ID() > targetMigration.ID() {
		return nil, errors.Wrapf(ErrCurrentMigrationMoreRecent, "current %s, target %s", current.ID(), targetMigration.ID())
	}

	// Build plan
	lst := list[currentMigrationIndex+1:]
	plan := make(Plan, len(lst))
	for i, m := range lst {
		plan[i] = &Action{
			Action:    ActionTypeDo,
			Migration: m,
		}
	}

	return plan, nil
}
