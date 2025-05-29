package check

import (
	"context"
	"database/sql"
	"fmt"
	"os"
)

func readSQLiteSequence(ctx context.Context, db *sql.DB, name string) (int64, error) {
	const query = "SELECT seq FROM sqlite_sequence WHERE name = ?"
	var seq int64

	row := db.QueryRowContext(ctx, query, name)
	err := row.Scan(&seq)
	return seq, err
}

func CheckSQLiteSequence(ctx context.Context, db4 *sql.DB, db5 *sql.DB) bool {
	names := [...]string{
		"comps_environment_group",
		"comps_group_package",
		"trans",
		"trans_item",
	}

	for _, name := range names {
		seq4, err := readSQLiteSequence(ctx, db4, name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Read DNF 4 sqlite_sequence %s error: %v\n",
				name, err)
			return false
		}

		seq5, err := readSQLiteSequence(ctx, db5, name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Read DNF 5 sqlite_sequence %s error: %v\n",
				name, err)
			return false
		}

		if seq4 != seq5 {
			fmt.Fprintf(os.Stderr, "sqlite_sequence %s error: %d (DNF 4) ≠ %d (DNF 5)\n",
				name, seq4, seq5)
			return false
		}
	}
	return true
}
