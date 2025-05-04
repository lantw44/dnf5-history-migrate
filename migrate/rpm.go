package migrate

import (
	"context"
	"database/sql"
)

type dnf4RPM struct {
	// CREATE TABLE rpm (
	item_id int64  // INTEGER UNIQUE NOT NULL,
	name    string // TEXT NOT NULL,
	epoch   int32  // INTEGER NOT NULL,        /* empty epoch is stored as 0 */
	version string // TEXT NOT NULL,
	release string // TEXT NOT NULL,
	arch    string // TEXT NOT NULL,
	//     FOREIGN KEY(item_id) REFERENCES item(id),
	//     CONSTRAINT rpm_unique_nevra UNIQUE (name, epoch, version, release, arch)
	// );
}

type dnf5RPM struct {
	// CREATE TABLE "rpm" (
	item_id int64  // INTEGER NOT NULL UNIQUE,
	name_id int64  // INTEGER NOT NULL,
	epoch   int32  // INTEGER NOT NULL,        /* empty epoch is stored as 0 */
	version string // TEXT NOT NULL,
	release string // TEXT NOT NULL,
	arch_id int64  // INTEGER NOT NULL,
	//     FOREIGN KEY("item_id") REFERENCES "item"("id"),
	//     FOREIGN KEY("name_id") REFERENCES "pkg_name"("id"),
	//     FOREIGN KEY("arch_id") REFERENCES "arch"("id"),
	//     CONSTRAINT "rpm_unique_nevra" UNIQUE ("name_id", "epoch", "version", "release", "arch_id")
	// );
}

func insertPkgName(ctx context.Context, tx5 *sql.Tx, name string) (int64, error) {
	_, err := tx5.ExecContext(ctx, `INSERT INTO pkg_name (name)
		VALUES (?) ON CONFLICT DO NOTHING`, name)
	if err != nil {
		return 0, FuncError{"tx5.ExecContext", err}
	}

	var id int64
	row := tx5.QueryRowContext(ctx, `SELECT id FROM pkg_name
		WHERE name = ?`, name)
	err = row.Scan(&id)
	if err != nil {
		return 0, FuncError{"tx5.QueryRowContext", err}
	}
	return id, nil
}

func insertArch(ctx context.Context, tx5 *sql.Tx, arch string) (int64, error) {
	_, err := tx5.ExecContext(ctx, `INSERT INTO arch (name)
		VALUES (?) ON CONFLICT DO NOTHING`, arch)
	if err != nil {
		return 0, FuncError{"tx5.ExecContext", err}
	}

	var id int64
	row := tx5.QueryRowContext(ctx, `SELECT id FROM arch
		WHERE name = ?`, arch)
	err = row.Scan(&id)
	if err != nil {
		return 0, FuncError{"tx5.QueryRowContext", err}
	}
	return id, nil
}

func migrateRPM(ctx context.Context, db4 *sql.DB, db5 *sql.DB) error {
	rows, err := db4.QueryContext(ctx, `SELECT
		item_id, name,
		epoch, version, release, arch
	FROM rpm ORDER BY item_id ASC`)
	if err != nil {
		return FuncError{"db4.QueryContext", err}
	}
	defer rows.Close()

	return transaction(ctx, db5, func(tx5 *sql.Tx) error {
		for rows.Next() {
			r4 := dnf4RPM{}
			err := rows.Scan(&r4.item_id, &r4.name,
				&r4.epoch, &r4.version, &r4.release, &r4.arch)
			if err != nil {
				return FuncError{"rows.Scan", err}
			}

			// commit 7537165036808bca5afbd5774f0e1cf0cc30f790
			// HistoryDB: Create table with pkg names, use them as foreign keys in rpm table
			name, err := insertPkgName(ctx, tx5, r4.name)
			if err != nil {
				return FuncError{"insertPkgName", err}
			}
			// commit 49f2799a63471c06c565e501030e567342d0c07f
			// HistoryDB: Create table with archs, use them as foreign keys in rpm table
			arch, err := insertArch(ctx, tx5, r4.arch)
			if err != nil {
				return FuncError{"insertArch", err}
			}
			r5 := dnf5RPM{
				item_id: r4.item_id,
				name_id: name,
				epoch:   r4.epoch,
				version: r4.version,
				release: r4.release,
				arch_id: arch,
			}
			_, err = tx5.ExecContext(ctx, `INSERT INTO rpm (
				item_id, name_id,
				epoch, version, release, arch_id
			) VALUES (?, ?, ?, ?, ?, ?)`,
				r5.item_id, r5.name_id,
				r5.epoch, r5.version, r5.release, r5.arch_id)
			if err != nil {
				return FuncError{"tx5.ExecContext", err}
			}
		}
		return rows.Err()
	})
}
