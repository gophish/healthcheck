package db

import (
	"bitbucket.org/liamstask/goose/lib/goose"
	log "github.com/gophish/gophish/logger"
	"github.com/gophish/healthcheck/config"
	"github.com/jinzhu/gorm"
)

var db *gorm.DB

func chooseDBDriver(name, openStr string) goose.DBDriver {
	d := goose.DBDriver{Name: name, OpenStr: openStr}

	switch name {
	case "mysql":
		d.Import = "github.com/go-sql-driver/mysql"
		d.Dialect = &goose.MySqlDialect{}

	// Default database is sqlite3
	default:
		d.Import = "github.com/mattn/go-sqlite3"
		d.Dialect = &goose.Sqlite3Dialect{}
	}

	return d
}

func Setup() error {
	// Setup the goose configuration
	log.Infof("%#v", config.Config)
	migrateConf := &goose.DBConf{
		MigrationsDir: config.Config.MigrationsPath,
		Env:           "production",
		Driver:        chooseDBDriver(config.Config.DBName, config.Config.DBPath),
	}
	// Get the latest possible migration
	latest, err := goose.GetMostRecentDBVersion(migrateConf.MigrationsDir)
	if err != nil {
		log.Error(err)
		return err
	}
	// Open our database connection
	db, err = gorm.Open(config.Config.DBName, config.Config.DBPath)
	db.LogMode(false)
	db.SetLogger(log.Logger)
	db.DB().SetMaxOpenConns(1)
	if err != nil {
		log.Error(err)
		return err
	}
	// Migrate up to the latest version
	err = goose.RunMigrationsOnDb(migrateConf, migrateConf.MigrationsDir, latest, db.DB())
	if err != nil {
		log.Error(err)
		return err
	}
	return nil
}
