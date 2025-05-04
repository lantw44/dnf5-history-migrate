package migrate

import (
	"context"
	"database/sql"
)

type dnf4CompsEnvironment struct {
	// CREATE TABLE comps_environment (
	item_id         int64  // INTEGER UNIQUE NOT NULL,
	environmentid   string // TEXT NOT NULL,
	name            string // TEXT NOT NULL,
	translated_name string // TEXT NOT NULL,
	pkg_types       int32  // INTEGER NOT NULL,
	//     FOREIGN KEY(item_id) REFERENCES item(id)
	// );
}

type dnf5CompsEnvironment struct {
	// CREATE TABLE "comps_environment" (
	item_id         int64  // INTEGER NOT NULL UNIQUE,
	environmentid   string // TEXT NOT NULL,
	name            string // TEXT NOT NULL,
	translated_name string // TEXT NOT NULL,
	pkg_types       int32  // INTEGER NOT NULL,
	//     FOREIGN KEY("item_id") REFERENCES "item"("id")
	// );
}

func migrateCompsEnvironment(ctx context.Context, db4 *sql.DB, db5 *sql.DB) error {
	rows, err := db4.QueryContext(ctx, `SELECT
		item_id, environmentid,
		name, translated_name, pkg_types
	FROM comps_environment ORDER BY item_id ASC`)
	if err != nil {
		return FuncError{"db4.QueryContext", err}
	}
	defer rows.Close()

	return transaction(ctx, db5, func(tx5 *sql.Tx) error {
		for rows.Next() {
			r4 := dnf4CompsEnvironment{}
			err := rows.Scan(&r4.item_id, &r4.environmentid,
				&r4.name, &r4.translated_name, &r4.pkg_types)
			if err != nil {
				return FuncError{"rows.Scan", err}
			}

			r5 := dnf5CompsEnvironment(r4)
			_, err = tx5.ExecContext(ctx, `INSERT INTO comps_environment (
				item_id, environmentid,
				name, translated_name, pkg_types
			) VALUES (?, ?, ?, ?, ?)`,
				r5.item_id, r5.environmentid,
				r5.name, r5.translated_name, r5.pkg_types)
			if err != nil {
				return FuncError{"tx5.ExecContext", err}
			}
		}
		return rows.Err()
	})
}
