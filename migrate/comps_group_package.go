package migrate

import (
	"context"
	"database/sql"
)

type dnf4CompsGroupPackage struct {
	// CREATE TABLE comps_group_package (
	id        int64  // INTEGER PRIMARY KEY AUTOINCREMENT,
	group_id  int64  // INTEGER NOT NULL,
	name      string // TEXT NOT NULL,
	installed int32  // INTEGER NOT NULL,
	pkg_type  int32  // INTEGER NOT NULL,
	//     FOREIGN KEY(group_id) REFERENCES comps_group(item_id),
	//     CONSTRAINT comps_group_package_unique_name UNIQUE (group_id, name)
	// );
}

type dnf5CompsGroupPackage struct {
	// CREATE TABLE "comps_group_package" (
	id        int64 // INTEGER,
	group_id  int64 // INTEGER NOT NULL,
	name_id   int64 // INTEGER NOT NULL,
	installed int32 // INTEGER NOT NULL,
	pkg_type  int32 // INTEGER NOT NULL,
	//     FOREIGN KEY("group_id") REFERENCES "comps_group"("item_id"),
	//     FOREIGN KEY("name_id") REFERENCES "pkg_name"("id"),
	//     CONSTRAINT "comps_group_package_unique_name" UNIQUE ("group_id", "name_id"),
	//     PRIMARY KEY("id" AUTOINCREMENT)
	// );
}

func migrateCompsGroupPackage(ctx context.Context, db4 *sql.DB, db5 *sql.DB) error {
	rows, err := db4.QueryContext(ctx, `SELECT
		id, group_id, name,
		installed, pkg_type
	FROM comps_group_package ORDER BY id ASC`)
	if err != nil {
		return FuncError{"db4.QueryContext", err}
	}
	defer rows.Close()

	return transaction(ctx, db5, func(tx5 *sql.Tx) error {
		for rows.Next() {
			r4 := dnf4CompsGroupPackage{}
			err := rows.Scan(&r4.id, &r4.group_id, &r4.name,
				&r4.installed, &r4.pkg_type)
			if err != nil {
				return FuncError{"rows.Scan", err}
			}

			// commit f958137cd4d081b47cf0fc326e34c175174bb672
			// HistoryDB: comps_group_package: Refer to pkg name from the pkg_names table
			name, err := insertPkgName(ctx, tx5, r4.name)
			if err != nil {
				return FuncError{"insertPkgName", err}
			}
			r5 := dnf5CompsGroupPackage{
				id:        r4.id,
				group_id:  r4.group_id,
				name_id:   name,
				installed: r4.installed,
				pkg_type:  r4.pkg_type,
			}
			_, err = tx5.ExecContext(ctx, `INSERT INTO comps_group_package (
				id, group_id, name_id,
				installed, pkg_type
			) VALUES (?, ?, ?, ?, ?)`,
				r5.id, r5.group_id, r5.name_id,
				r5.installed, r5.pkg_type)
			if err != nil {
				return FuncError{"tx5.ExecContext", err}
			}
		}
		return rows.Err()
	})
}
