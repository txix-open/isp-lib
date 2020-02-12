package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/go-pg/pg/v9"
	"github.com/integration-system/isp-lib/v2/structure"
)

func NewDbConnection(config structure.DBConfiguration) (*pg.DB, error) {
	pdb := pg.Connect(&pg.Options{
		User:               config.Username,
		Password:           config.Password,
		Database:           config.Database,
		Addr:               config.Address + ":" + config.Port,
		MaxRetries:         5,
		IdleTimeout:        time.Duration(15) * time.Minute,
		IdleCheckFrequency: time.Duration(30) * time.Second,
		PoolSize:           config.PoolSize,
	})

	var n time.Time
	_, err := pdb.QueryOne(pg.Scan(&n), "SELECT now()")

	return pdb, err
}

func openSqlConn(config structure.DBConfiguration, schema string) (*sql.DB, error) {
	cs := fmt.Sprintf(
		"postgres://%s:%s/%s?search_path=%s,public&sslmode=disable&user=%s&password=%s",
		config.Address,
		config.Port,
		config.Database,
		schema,
		config.Username,
		config.Password,
	)
	return sql.Open("postgres", cs)
}
