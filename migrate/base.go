package migrate

import (
	"context"
	"database/sql"
	"fmt"
)

type EnumError struct {
	name  string
	value int32
}

func (e EnumError) Error() string {
	return fmt.Sprintf("unknown %s value: %d", e.name, e.value)
}

type FuncError struct {
	name string
	err  error
}

func (e FuncError) Error() string {
	return fmt.Sprintf("call %s error: %v", e.name, e.err)
}

func (e FuncError) Unwrap() error {
	return e.err
}

func transaction(ctx context.Context, db *sql.DB, f func(*sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	err = f(tx)
	if err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}
