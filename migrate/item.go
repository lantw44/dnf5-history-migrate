package migrate

import (
	"context"
	"database/sql"
)

type dnf4Item struct {
	// CREATE TABLE item (
	id        int64 // INTEGER PRIMARY KEY,
	item_type int32 // INTEGER NOT NULL /* (enum) 1: rpm, 2: group, 3: env ...*/
	// );
}

type dnf5Item struct {
	// CREATE TABLE "item" (
	id int64 // INTEGER,
	// PRIMARY KEY("id")
	// );
}

func migrateItem(ctx context.Context, db4 *sql.DB, db5 *sql.DB) error {
	rows, err := db4.QueryContext(ctx, `SELECT
		id, item_type
	FROM item ORDER BY id ASC`)
	if err != nil {
		return FuncError{"db4.QueryContext", err}
	}
	defer rows.Close()

	return transaction(ctx, db5, func(tx5 *sql.Tx) error {
		for rows.Next() {
			r4 := dnf4Item{}
			err := rows.Scan(&r4.id, &r4.item_type)
			if err != nil {
				return FuncError{"rows.Scan", err}
			}

			// commit 5b492afbae39eb7cf7d2b11e677d460f756349fa
			// libdnf/transaction: Drop TransactionItemType
			r5 := dnf5Item{id: r4.id}
			_, err = tx5.ExecContext(ctx, `INSERT INTO item (
				id
			) VALUES (?)`, r5.id)
			if err != nil {
				return FuncError{"tx5.ExecContext", err}
			}
		}
		return rows.Err()
	})
}
