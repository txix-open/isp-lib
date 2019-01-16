package database

import (
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"github.com/jinzhu/inflection"
	_ "github.com/lib/pq"
	"github.com/pressly/goose"
	"github.com/integration-system/isp-lib/logger"
	"github.com/integration-system/isp-lib/utils"
	"os"
	"path"
	"strings"
)

const (
	SchemaParam = "db_schema"
)

var (
	dbManagerInstance         = DBManager{isInitialized: false}
	dbReadOnlyManagerInstance = DBManager{isInitialized: false}
	ormDbTypeName             = "orm.DB"
)

type DBConfiguration struct {
	Address      string `valid:"required~Required"`
	Schema       string `valid:"required~Required"`
	Database     string `valid:"required~Required"`
	Port         string `valid:"required~Required"`
	Username     string
	Password     string
	PoolSize     int
	CreateSchema bool
}

type DBManager struct {
	Db            *pg.DB
	isInitialized bool
}

func GetDBManager() *DBManager {
	if !dbManagerInstance.isInitialized {
		logger.Fatal("DbManager isn't init, call first the \"initDb\" method")
	}
	return &dbManagerInstance
}

func GetReadOnlyDBManager() *DBManager {
	if !dbReadOnlyManagerInstance.isInitialized {
		logger.Fatal("ReadOnlyDBManager isn't init, call first the \"initReadOnlyDb\" method")
	}
	return &dbReadOnlyManagerInstance
}

func Close() {
	if dbManagerInstance.Db != nil {
		if err := dbManagerInstance.Db.Close(); err != nil {
			logger.Warn(err)
		}
	}
}

func CloseReadOnly() {
	if dbReadOnlyManagerInstance.Db != nil {
		if err := dbReadOnlyManagerInstance.Db.Close(); err != nil {
			logger.Warn(err)
		}
	}
}

func InitDb(config DBConfiguration) {
	if config.CreateSchema {
		if err := ensureSchemaExists(config); err != nil {
			logger.Fatal(err)
		}
	}

	if err := ensureMigrations(config); err != nil {
		logger.Fatal(err)
	}

	Close()

	pdb, err := NewDbConnection(config)
	if err != nil {
		logger.Fatal(err)
	}

	dbManagerInstance = DBManager{Db: pdb, isInitialized: true}
}

func InitDbWithSchema(config DBConfiguration, schema string) {
	InitDb(config)
	dbManagerInstance.Db = withSchema(dbManagerInstance.Db, schema)
	orm.SetTableNameInflector(func(s string) string {
		return schema + "." + inflection.Plural(s)
	})
}

func InitDbWithCurrentSchema(config DBConfiguration) {
	InitDb(config)
	dbManagerInstance.Db = withSchema(dbManagerInstance.Db, config.Schema)
	orm.SetTableNameInflector(func(s string) string {
		return config.Schema + "." + inflection.Plural(s)
	})
}

func InitDbV2(config DBConfiguration, callback func(db *DBManager)) {
	InitDb(config)
	callback(GetDBManager())
}

func InitDbV2WithSchema(config DBConfiguration, schema string, callback func(db *DBManager)) {
	InitDb(config)
	dbManagerInstance.Db = withSchema(dbManagerInstance.Db, schema)
	orm.SetTableNameInflector(func(s string) string {
		return schema + "." + inflection.Plural(s)
	})
	callback(GetDBManager())
}

func initReadOnlyDb(config DBConfiguration) {
	pdb, err := NewDbConnection(config)
	if err != nil {
		logger.Fatal(err)
	}
	dbReadOnlyManagerInstance = DBManager{Db: pdb, isInitialized: true}
}

func InitReadOnlyDbWithSchema(config DBConfiguration, schema string) {
	initReadOnlyDb(config)
	dbReadOnlyManagerInstance.Db = withSchema(dbReadOnlyManagerInstance.Db, schema)
	orm.SetTableNameInflector(func(s string) string {
		return config.Schema + "." + inflection.Plural(s)
	})
}

func InitReadOnlyDbWithCurrentSchema(config DBConfiguration) {
	initReadOnlyDb(config)
	dbReadOnlyManagerInstance.Db = withSchema(dbReadOnlyManagerInstance.Db, config.Schema)
	orm.SetTableNameInflector(func(s string) string {
		return config.Schema + "." + inflection.Plural(s)
	})
}

func InitReadOnlyDbV2(config DBConfiguration, callback func(db *DBManager)) {
	initReadOnlyDb(config)
	callback(GetReadOnlyDBManager())
}

func InitReadOnlyDbV2WithSchema(config DBConfiguration, schema string, callback func(db *DBManager)) {
	initReadOnlyDb(config)
	dbReadOnlyManagerInstance.Db = withSchema(dbReadOnlyManagerInstance.Db, schema)
	orm.SetTableNameInflector(func(s string) string {
		return schema + "." + inflection.Plural(s)
	})
	callback(GetReadOnlyDBManager())
}

func ResolveMigrationsDirectrory() string {
	ex, _ := os.Executable()
	migrationDir := "migrations"
	if !utils.DEV {
		// _, filename, _, _ := runtime.Caller(0)
		migrationDir = path.Dir(ex) + "/" + migrationDir
	}
	if utils.EnvMigrationPath != "" {
		if strings.HasSuffix(utils.EnvMigrationPath, "/") {
			migrationDir = utils.EnvMigrationPath + migrationDir
		} else {
			migrationDir = utils.EnvMigrationPath + "/" + migrationDir
		}
	}
	return migrationDir
}

func ensureSchemaExists(config DBConfiguration) error {
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

func ensureMigrations(config DBConfiguration) error {
	db, err := openSqlConn(config, config.Schema)
	defer func() {
		if db != nil {
			db.Close()
		}
	}()
	if err != nil {
		return err
	}

	migrationDir := ResolveMigrationsDirectrory()

	if _, err := os.Stat(migrationDir); !os.IsNotExist(err) {
		goose.Version(db, migrationDir)
		goose.Status(db, migrationDir)
		if err := goose.Run("up", db, migrationDir); err != nil {
			return err
		}
	} else {
		logger.Infof("Migration directory is not exists: %s", migrationDir)
	}

	return nil
}

func withSchema(pdb *pg.DB, schema string) *pg.DB {
	return pdb.WithParam(SchemaParam, pg.F(schema))
}
