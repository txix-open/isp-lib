package database

import (
	"github.com/go-pg/pg"
	"github.com/integration-system/isp-lib/structure"
	"github.com/integration-system/isp-lib/utils"
	_ "github.com/lib/pq"
	"github.com/pressly/goose"
	"os"
	"path"
	"strings"
)

const (
	SchemaParam = "db_schema"
)

func ResolveMigrationsDirectory() string {
	ex, _ := os.Executable()
	migrationDir := "migrations"
	if !utils.DEV {
		// _, filename, _, _ := runtime.Caller(0)
		migrationDir = path.Dir(ex) + "/" + migrationDir
	}
	if utils.EnvMigrationPath != "" {
		if strings.HasPrefix(utils.EnvMigrationPath, "/") {
			migrationDir = utils.EnvMigrationPath
		} else {
			migrationDir = path.Dir(ex) + "/" + utils.EnvMigrationPath
		}
	}
	return migrationDir
}

func ensureSchemaExists(config structure.DBConfiguration) error {
	db, err := openSqlConn(config, "public")
	defer func() {
		if db != nil {
			db.Close()
		}
	}()
	if err != nil {
		return err
	}

	_, err = db.Exec("CREATE SCHEMA IF NOT EXISTS " + config.Schema)
	return err
}

func ensureMigrations(config structure.DBConfiguration) error {
	db, err := openSqlConn(config, config.Schema)
	defer func() {
		if db != nil {
			db.Close()
		}
	}()
	if err != nil {
		return err
	}

	migrationDir := ResolveMigrationsDirectory()

	if _, err := os.Stat(migrationDir); !os.IsNotExist(err) {
		goose.Version(db, migrationDir)
		goose.Status(db, migrationDir)
		if err := goose.Run("up", db, migrationDir); err != nil {
			return err
		}
	} else {
		return err
	}

	return nil
}

func withSchema(pdb *pg.DB, schema string) *pg.DB {
	if pdb != nil {
		return pdb.WithParam(SchemaParam, pg.F(schema))
	}
	return nil
}
