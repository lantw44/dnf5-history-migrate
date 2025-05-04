package migrate

import (
	"context"
	"database/sql"
)

type dnf4CompsGroup struct {
	// CREATE TABLE comps_group (
	item_id         int64  // INTEGER UNIQUE NOT NULL,
	groupid         string // TEXT NOT NULL,
	name            string // TEXT NOT NULL,
	translated_name string // TEXT NOT NULL,
	pkg_types       int32  // INTEGER NOT NULL,
	//     FOREIGN KEY(item_id) REFERENCES item(id)
	// );
}

type dnf5CompsGroup struct {
	// CREATE TABLE "comps_group" (
	item_id         int64  // INTEGER NOT NULL UNIQUE,
	groupid         string // TEXT NOT NULL,
	name            string // TEXT NOT NULL,
	translated_name string // TEXT NOT NULL,
	pkg_types       int32  // INTEGER NOT NULL,
	//     FOREIGN KEY("item_id") REFERENCES "item"("id")
	// );
}

func migrateCompsGroup(ctx context.Context, db4 *sql.DB, db5 *sql.DB) error {
	rows, err := db4.QueryContext(ctx, `SELECT
		item_id, groupid,
		name, translated_name, pkg_types
	FROM comps_group ORDER BY item_id ASC`)
	if err != nil {
		return FuncError{"db4.QueryContext", err}
	}
	defer rows.Close()

	return transaction(ctx, db5, func(tx5 *sql.Tx) error {
		for rows.Next() {
			r4 := dnf4CompsGroup{}
			err := rows.Scan(&r4.item_id, &r4.groupid,
				&r4.name, &r4.translated_name, &r4.pkg_types)
			if err != nil {
				return FuncError{"rows.Scan", err}
			}

			r5 := dnf5CompsGroup(r4)
			_, err = tx5.ExecContext(ctx, `INSERT INTO comps_group (
				item_id, groupid,
				name, translated_name, pkg_types
			) VALUES (?, ?, ?, ?, ?)`,
				r5.item_id, r5.groupid,
				r5.name, r5.translated_name, r5.pkg_types)
			if err != nil {
				return FuncError{"tx5.ExecContext", err}
			}
		}
		return rows.Err()
	})
}
