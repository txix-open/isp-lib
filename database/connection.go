package database

import (
	"database/sql"
	"github.com/go-pg/pg"
	"github.com/integration-system/isp-lib/logger"
	"github.com/integration-system/isp-lib/structure"
	"time"
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
	if err == nil {
		logger.Infof("Database server time: %d", n)
	}

	return pdb, err
}

func openSqlConn(config structure.DBConfiguration, schema string) (*sql.DB, error) {
	return sql.Open("postgres", "postgres://"+
		config.Address+":"+config.Port+
		"/"+config.Database+
		"?search_path="+schema+"&sslmode=disable&user="+
		config.Username+"&password="+config.Password)
}
