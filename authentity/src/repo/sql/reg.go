package sql

import (
	"context"
	"database/sql"
	"github.com/mattn/go-sqlite3"
	"log"
)

func RegisterHandlerToDB(db *sql.DB, handler any) {
	conn, err := db.Conn(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	err = conn.Raw(func(driverConn interface{}) error {
		sqliteConn := driverConn.(*sqlite3.SQLiteConn)
		defer sqliteConn.Close()

		return sqliteConn.RegisterFunc("my_custom_function", handler, true)
	})
	if err != nil {
		log.Fatal(err)
	}
}
