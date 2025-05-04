package migrate

import (
	"context"
	"database/sql"
)

type dnf4CompsEnvironmentGroup struct {
	// CREATE TABLE comps_environment_group (
	id             int64  // INTEGER PRIMARY KEY AUTOINCREMENT,
	environment_id int64  // INTEGER NOT NULL,
	groupid        string // TEXT NOT NULL,
	installed      int32  // INTEGER NOT NULL,
	group_type     int32  // INTEGER NOT NULL,
	//     FOREIGN KEY(environment_id) REFERENCES comps_environment(item_id),
	//     CONSTRAINT comps_environment_group_unique_groupid UNIQUE (environment_id, groupid)
	// );
}
type dnf5CompsEnvironmentGroup struct {
	// CREATE TABLE "comps_environment_group" (
	id             int64  // INTEGER,
	environment_id int64  // INTEGER NOT NULL,
	groupid        string // TEXT NOT NULL,
	installed      int32  // INTEGER NOT NULL,
	group_type     int32  // INTEGER NOT NULL,
	//     FOREIGN KEY("environment_id") REFERENCES "comps_environment"("item_id"),
	//     CONSTRAINT "comps_environment_group_unique_groupid" UNIQUE ("environment_id", "groupid"),
	//     PRIMARY KEY("id" AUTOINCREMENT)
	// );
}

func migrateCompsEnvironmentGroup(ctx context.Context, db4 *sql.DB, db5 *sql.DB) error {
	rows, err := db4.QueryContext(ctx, `SELECT
		id, environment_id, groupid,
		installed, group_type
	FROM comps_environment_group ORDER BY id ASC`)
	if err != nil {
		return FuncError{"db4.QueryContext", err}
	}
	defer rows.Close()

	return transaction(ctx, db5, func(tx5 *sql.Tx) error {
		for rows.Next() {
			r4 := dnf4CompsEnvironmentGroup{}
			err := rows.Scan(
				&r4.id, &r4.environment_id, &r4.groupid,
				&r4.installed, &r4.group_type)
			if err != nil {
				return FuncError{"rows.Scan", err}
			}

			r5 := dnf5CompsEnvironmentGroup(r4)
			_, err = tx5.ExecContext(ctx, `INSERT INTO comps_environment_group (
				id, environment_id, groupid,
				installed, group_type
			) VALUES (?, ?, ?, ?, ?)`,
				r5.id, r5.environment_id, r5.groupid,
				r5.installed, r5.group_type)
			if err != nil {
				return FuncError{"tx5.ExecContext", err}
			}
		}
		return rows.Err()
	})
}
