package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os"

	"github.com/lantw44/dnf5-history-migrate/check"
	"github.com/lantw44/dnf5-history-migrate/migrate"
	_ "modernc.org/sqlite"
)

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

	if !check.CheckConfig4(ctx, db4) {
		return 2
	}

	db5, err := sql.Open("sqlite", path5)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Open DNF 5 database error: %v\n", err)
		return 1
	}
	defer db5.Close()

	if !check.CheckConfig5(ctx, db5) {
		return 2
	}

	err = migrate.Migrate(ctx, db4, db5)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 3
	}

	if !check.CheckSQLiteSequence(ctx, db4, db5) {
		return 4
	}
	return 0
}

func main() {
	os.Exit(main2())
}
