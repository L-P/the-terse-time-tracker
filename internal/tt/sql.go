package tt

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"log"
	"strings"
)

type transactionCallback func(*sql.Tx) error

// transaction runs the given callback in a SQL transaction and either COMMIT
// if the returned error is nil, or ROLLBACK if the error is non-nil.
// All interactions with the DB should go through this function.
func (tt *TT) transaction(cb transactionCallback) error {
	tx, err := tt.db.Begin()
	if err != nil {
		return ErrDatabase(err.Error())
	}

	if err := cb(tx); err != nil {
		if err2 := tx.Rollback(); err2 != nil {
			return ErrDatabase(fmt.Sprintf("rollback error: %s\noriginal error: %s", err2, err))
		}

		return err
	}

	if err := tx.Commit(); err != nil {
		return ErrDatabase(err.Error())
	}

	return nil
}

func exec(tx *sql.Tx, query string, params ...interface{}) error {
	if _, err := tx.Exec(query, params...); err != nil {
		return ErrBadQuery{err, query, params}
	}

	return nil
}

func debugPrintQuery(query string, params ...interface{}) { // nolint:deadcode,unused
	for _, v := range params {
		var str string
		if v == nil {
			str = "NULL"
		} else if valuer, ok := v.(driver.Valuer); ok {
			value, _ := valuer.Value()
			if value != nil {
				str = fmt.Sprintf(`'%v'`, value)
			} else {
				str = "NULL"
			}
		} else {
			str = fmt.Sprintf(`'%s'`, v)
		}

		query = strings.Replace(query, "?", str, 1)
	}

	log.Printf("debug: %s", query)
}
