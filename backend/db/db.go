package db

import (
	"fmt"
	"log"
	"moneropot/monerorpc"
	"moneropot/util"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/gabstv/httpdigest"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

var (
	dbLock       sync.Mutex
	dbx          *sqlx.DB
	dbPath       string
	CurrentPrice uint64

	DateFormat      = "2006-01-02"
	DateTimeFormat  = "2006-01-02 15:04:05"
	SDateTimeFormat = "2006-01-02T15:04:05.000Z"
)

func Init() {
	if err := os.MkdirAll(util.Config.DataPath, 0755); err != nil {
		log.Fatal(err)
	}
	dbPath = filepath.Join(util.Config.DataPath, util.Config.DbName)
	if util.Config.DbName == ":memory:" {
		dbPath = util.Config.DbName
	}
	Wallet = monerorpc.New(monerorpc.Config{
		Address:   util.Config.RpcAddress,
		Transport: httpdigest.New(util.Config.RpcUser, util.Config.RpcPass),
	})
	Daemon = monerorpc.New(monerorpc.Config{
		Address:   util.Config.DaemonAddress,
		Transport: httpdigest.New(util.Config.DaemonUser, util.Config.DaemonPass),
	})
	// backup if already exists on every update then every 24 hours
	doBackup()
	MustDB()
	if err := SetCurrentPrice(); err != nil {
		log.Fatal(err)
	}
}

func MustDB() *sqlx.DB {
	db, err := GetDB()
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func GetDB() (*sqlx.DB, error) {
	dbLock.Lock()
	defer dbLock.Unlock()
	if dbx != nil {
		return dbx, nil
	}
	log.Printf("Opening %s", dbPath)
	db, err := sqlx.Open("sqlite3", dbPath)
	db.SetMaxOpenConns(1)
	if err != nil {
		return nil, fmt.Errorf("GetDB sqlx.Open error: %v", err)
	}
	dbx = db
	tdb := len(dbMigrations)
	r := db.QueryRow(`SELECT value FROM metadata WHERE key = 'db_version'`)
	var (
		ver     string
		version int
	)
	if err := r.Scan(&ver); err != nil {
		if err.Error() == "no such table: metadata" {
			version = 0
		} else {
			return nil, fmt.Errorf("GetDB r.Scan error: %v", err)
		}
	}
	if ver != "" {
		version, _ = strconv.Atoi(ver)
	}
	if tdb > version {
		tx, err := db.Begin()
		if err != nil {
			return nil, fmt.Errorf("GetDB db.Exec behin migration error: %v", err)
		}
		for i := version; i < tdb; i++ {
			_, err = tx.Exec(dbMigrations[i])
			if err != nil {
				log.Println("RollbackError: migration ", tx.Rollback())
				return nil, fmt.Errorf("GetDB tx.Exec migrations error: %d -> %v", i, err)
			}
		}
		if _, err := tx.Exec(`UPDATE metadata SET value = $1 WHERE key = 'db_version'`, strconv.Itoa(tdb)); err != nil {
			log.Println("RollbackError: version update ", tx.Rollback())
			return nil, fmt.Errorf("GetDB tx.Exec set version error: %v", err)
		}
		if err := tx.Commit(); err != nil {
			return nil, fmt.Errorf("GetDB tx.Commit error: %v", err)
		}
		log.Printf("Migrated db to %d", tdb)
	}
	return db, nil
}
