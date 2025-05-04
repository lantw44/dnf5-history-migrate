package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os"

	"github.com/lantw44/dnf5-history-migrate/migrate"
	_ "modernc.org/sqlite"
)

func checkConfig4(ctx context.Context, db *sql.DB) bool {
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

func checkConfig5(ctx context.Context, db *sql.DB) bool {
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

func main2() int {
	flag.Parse()
	path4 := flag.Arg(0)
	path5 := flag.Arg(1)

	fmt.Printf("Source DNF 4 database: %s\n", path4)
	fmt.Printf("Target DNF 5 database: %s\n", path5)

	ctx := context.Background()

	db4, err := sql.Open("sqlite", path4)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Open DNF 4 database error: %v\n", err)
		return 1
	}
	defer db4.Close()

	config4 := checkConfig4(ctx, db4)
	if !config4 {
		return 2
	}

	db5, err := sql.Open("sqlite", path5)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Open DNF 5 database error: %v\n", err)
		return 1
	}
	defer db5.Close()

	config5 := checkConfig5(ctx, db5)
	if !config5 {
		return 2
	}

	err = migrate.Migrate(ctx, db4, db5)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 3
	}
	return 0
}

func main() {
	os.Exit(main2())
}
