package migrate

import (
	"context"
	"database/sql"
	"fmt"
)

type step struct {
	name string
	run  func(context.Context, *sql.DB, *sql.DB) error
}

func Migrate(ctx context.Context, db4 *sql.DB, db5 *sql.DB) error {
	steps := []step{
		{"migrateTrans", migrateTrans},
		{"migrateRepo", migrateRepo},
		// commit 42beeae737f73618209b522e26dcd4c69c378c81
		// [libdnf] Remove "console_output" table from history database
		{"migrateItem", migrateItem},
		{"migrateTransItem", migrateTransItem},
		{"migrateItemReplacedBy", migrateItemReplacedBy},
		// commit 613add0f37fcf9d8a8058ce6e309d5bafa93465a
		// libdnf/transaction/db: Drop the trans_with table
		{"migrateRPM", migrateRPM},
		{"migrateCompsGroup", migrateCompsGroup},
		{"migrateCompsGroupPackage", migrateCompsGroupPackage},
		{"migrateCompsEnvironment", migrateCompsEnvironment},
		{"migrateCompsEnvironmentGroup", migrateCompsEnvironmentGroup},
	}

	for _, step := range steps {
		fmt.Printf("Run %s\n", step.name)
		if err := step.run(ctx, db4, db5); err != nil {
			return err
		}
	}
	return nil
}
