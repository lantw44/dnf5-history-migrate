package migrate

import (
	"context"
	"database/sql"
)

type dnf4ItemReplacedBy struct {
	// CREATE TABLE item_replaced_by ( /* M:N relationship between transaction items */
	trans_item_id    sql.Null[int64] // INTEGER REFERENCES trans_item(id),
	by_trans_item_id sql.Null[int64] // INTEGER REFERENCES trans_item(id),
	//     PRIMARY KEY (trans_item_id, by_trans_item_id)
	// );
}

type dnf5ItemReplacedBy struct {
	// CREATE TABLE "item_replaced_by" ( /* M:N relationship between transaction items */
	trans_item_id    sql.Null[int64] // INTEGER,
	by_trans_item_id sql.Null[int64] // INTEGER,
	//     PRIMARY KEY ("trans_item_id", "by_trans_item_id"),
	//     FOREIGN KEY("trans_item_id") REFERENCES "trans_item"("id"),
	//     FOREIGN KEY("by_trans_item_id") REFERENCES "trans_item"("id")
	// );
}

func migrateItemReplacedBy(ctx context.Context, db4 *sql.DB, db5 *sql.DB) error {
	rows, err := db4.QueryContext(ctx, `SELECT
		trans_item_id, by_trans_item_id
	FROM item_replaced_by ORDER BY trans_item_id ASC, by_trans_item_id ASC`)
	if err != nil {
		return FuncError{"db4.QueryContext", err}
	}
	defer rows.Close()

	return transaction(ctx, db5, func(tx5 *sql.Tx) error {
		for rows.Next() {
			r4 := dnf4ItemReplacedBy{}
			err := rows.Scan(&r4.trans_item_id, &r4.by_trans_item_id)
			if err != nil {
				return FuncError{"rows.Scan", err}
			}

			r5 := dnf5ItemReplacedBy(r4)
			_, err = tx5.ExecContext(ctx, `INSERT INTO item_replaced_by (
				trans_item_id, by_trans_item_id
			) VALUES (?, ?)`, r5.trans_item_id, r5.by_trans_item_id)
			if err != nil {
				return FuncError{"tx5.ExecContext", err}
			}
		}
		return rows.Err()
	})
}
