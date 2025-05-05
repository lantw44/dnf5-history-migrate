package check

import (
	"context"
	"database/sql"
	"fmt"
	"os"
)

func CheckConfig4(ctx context.Context, db *sql.DB) bool {
	const query = "SELECT value FROM config WHERE key = 'version'"
	const expected = "1.2"
	var actual string

	row := db.QueryRowContext(ctx, query)
	err := row.Scan(&actual)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Read DNF 4 database version error: %v\n", err)
		return false
	}

	if actual != expected {
		fmt.Fprintf(os.Stderr, "Bad DNF 4 database version %s (should be %s)\n",
			actual, expected)
		return false
	}
	return true
}

func CheckConfig5(ctx context.Context, db *sql.DB) bool {
	const query = "SELECT value FROM config WHERE key = 'version'"
	const expected = "1.1"
	var actual string

	row := db.QueryRowContext(ctx, query)
	err := row.Scan(&actual)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Read DNF 5 database version error: %v\n", err)
		return false
	}

	if actual != expected {
		fmt.Fprintf(os.Stderr, "Bad DNF 5 database version %s (should be %s)\n",
			actual, expected)
		return false
	}
	return true
}
