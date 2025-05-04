package migrate

import (
	"context"
	"database/sql"
)

type dnf4Repo struct {
	// CREATE TABLE repo (
	id     int64  // INTEGER PRIMARY KEY,
	repoid string // TEXT NOT NULL            /* repository ID aka 'repoid' */
	// );
}

type dnf5Repo struct {
	// CREATE TABLE "repo" (
	id     int64  // INTEGER,
	repoid string // TEXT NOT NULL,            /* repository ID aka 'repoid' */
	//     PRIMARY KEY("id")
	// );
}

func migrateRepo(ctx context.Context, db4 *sql.DB, db5 *sql.DB) error {
	rows, err := db4.QueryContext(ctx, `SELECT
		id, repoid
	FROM repo ORDER BY id ASC`)
	if err != nil {
		return FuncError{"db4.QueryContext", err}
	}
	defer rows.Close()

	return transaction(ctx, db5, func(tx5 *sql.Tx) error {
		for rows.Next() {
			r4 := dnf4Repo{}
			err := rows.Scan(&r4.id, &r4.repoid)
			if err != nil {
				return FuncError{"rows.Scan", err}
			}

			r5 := dnf5Repo(r4)
			_, err = tx5.ExecContext(ctx, `INSERT INTO repo (
				id, repoid
			) VALUES (?, ?)`, r5.id, r5.repoid)
			if err != nil {
				return FuncError{"tx5.ExecContext", err}
			}
		}
		return rows.Err()
	})
}
