package migrate

import (
	"context"
	"database/sql"
)

type dnf4TransState int32

const (
	dnf4TransStateUNKNOWN dnf4TransState = 0
	dnf4TransStateDONE    dnf4TransState = 1
	dnf4TransStateERROR   dnf4TransState = 2
)

type dnf4Trans struct {
	// CREATE TABLE trans (
	id                  int64            // INTEGER PRIMARY KEY AUTOINCREMENT,
	dt_begin            int64            // INTEGER NOT NULL,    /* (unix timestamp) date and time of transaction begin */
	dt_end              sql.Null[int64]  // INTEGER,             /* (unix timestamp) date and time of transaction end */
	rpmdb_version_begin sql.Null[string] // TEXT,
	rpmdb_version_end   sql.Null[string] // TEXT,
	releasever          string           // TEXT NOT NULL,       /* var: $releasever */
	user_id             int32            // INTEGER NOT NULL,    /* user ID (UID) */
	cmdline             sql.Null[string] // TEXT,                /* recorded command line (program, options, arguments) */
	state               dnf4TransState   // INTEGER NOT NULL     /* (enum) */
	// );
	// ALTER TABLE trans ADD
	comment sql.Null[string] // TEXT DEFAULT '';
}

type dnf5TransState int32

const (
	dnf5TransStateStarted dnf5TransState = 1
	dnf5TransStateOk      dnf5TransState = 2
	dnf5TransStateError   dnf5TransState = 3
)

type dnf5Trans struct {
	// CREATE TABLE "trans" (
	id                  int64            // INTEGER,
	dt_begin            int64            // INTEGER NOT NULL,    /* (unix timestamp) date and time of transaction begin */
	dt_end              sql.Null[int64]  // INTEGER,             /* (unix timestamp) date and time of transaction end */
	rpmdb_version_begin sql.Null[string] // TEXT,
	rpmdb_version_end   sql.Null[string] // TEXT,
	releasever          string           // TEXT NOT NULL,       /* var: $releasever */
	user_id             int32            // INTEGER NOT NULL,    /* user ID (UID) */
	description         sql.Null[string] // TEXT,                /* A description of the transaction (e.g. the CLI command being executed) */
	comment             sql.Null[string] // TEXT,                /* An arbitrary comment */
	state_id            dnf5TransState   // INTEGER,             /* (enum) */
	//     PRIMARY KEY("id" AUTOINCREMENT),
	//     FOREIGN KEY("state_id") REFERENCES "trans_state"("id")
	// );
}

func castTransState(s dnf4TransState) (dnf5TransState, error) {
	// commit d7affae879b7544f5eef3762d07bf15c7d7846f5
	// libdnf/transaction: Store TransactionState as enum in history db
	switch s {
	case dnf4TransStateUNKNOWN:
		return dnf5TransStateStarted, nil
	case dnf4TransStateDONE:
		return dnf5TransStateOk, nil
	case dnf4TransStateERROR:
		return dnf5TransStateError, nil
	default:
		return 0, EnumError{"dnf4TransState", int32(s)}
	}
}

func migrateTrans(ctx context.Context, db4 *sql.DB, db5 *sql.DB) error {
	rows, err := db4.QueryContext(ctx, `SELECT
		id, dt_begin, dt_end,
		rpmdb_version_begin, rpmdb_version_end,
		releasever, user_id, cmdline, state, comment
	FROM trans ORDER BY id ASC`)
	if err != nil {
		return FuncError{"db4.QueryContext", err}
	}
	defer rows.Close()

	return transaction(ctx, db5, func(tx5 *sql.Tx) error {
		for rows.Next() {
			r4 := dnf4Trans{}
			err := rows.Scan(&r4.id, &r4.dt_begin, &r4.dt_end,
				&r4.rpmdb_version_begin, &r4.rpmdb_version_end,
				&r4.releasever, &r4.user_id, &r4.cmdline, &r4.state, &r4.comment)
			if err != nil {
				return FuncError{"rows.Scan", err}
			}

			// commit 996be680e28a42c10a6d5255d36849db07cd5acb
			// libdnf/transaction/db: Rename cmdline to description, add comment
			state, err := castTransState(r4.state)
			if err != nil {
				return FuncError{"castTransState", err}
			}
			r5 := dnf5Trans{
				id:                  r4.id,
				dt_begin:            r4.dt_begin,
				dt_end:              r4.dt_end,
				rpmdb_version_begin: r4.rpmdb_version_begin,
				rpmdb_version_end:   r4.rpmdb_version_end,
				releasever:          r4.releasever,
				user_id:             r4.user_id,
				description:         r4.cmdline,
				comment:             r4.comment,
				state_id:            state,
			}

			_, err = tx5.ExecContext(ctx, `INSERT INTO trans (
				id, dt_begin, dt_end,
				rpmdb_version_begin, rpmdb_version_end,
				releasever, user_id, description, comment, state_id
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
				r5.id, r5.dt_begin, r5.dt_end,
				r5.rpmdb_version_begin, r5.rpmdb_version_end,
				r5.releasever, r5.user_id, r5.description, r5.comment, r5.state_id)
			if err != nil {
				return FuncError{"tx5.ExecContext", err}
			}
		}
		return rows.Err()
	})
}
