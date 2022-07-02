package database

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/plankton4/chat-app-server/cmd/server/config"
	"github.com/plankton4/chat-app-server/cmd/server/misc"
)

const (
	localAddr  = "localhost:3306"
	remoteAddr = config.RemoteMySQLAddress

	username = "plankton4"
	password = "plankton4_pass"
	dbname   = "chatapp"
)

var hostname string
var MainDB *sql.DB

func init() {
	if misc.IsConnectedToRemoteServer() {
		hostname = remoteAddr
	} else {
		hostname = localAddr
	}
	log.Println("Hostname ", hostname)
}

func OpenDatabase() *sql.DB {
	db, err := sql.Open("mysql", dsn(dbname))
	if err != nil {
		log.Println("Error when OpenDatabase!")
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		log.Println("Pinging database error ", err.Error())
	}

	MainDB = db

	//db.Exec("DROP TABLE IF EXISTS UserIDs;")
	//db.Exec("DROP TABLE IF EXISTS Users;")
	//db.Exec("DROP TABLE IF EXISTS SessionKeys")
	//db.Exec("DROP TABLE IF EXISTS FCMTokens")

	//
	// UserIDs
	//
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS UserIDs (
			id INT UNSIGNED AUTO_INCREMENT NOT NULL PRIMARY KEY
		);
	`)

	if err != nil {
		log.Println("OpenDatabase, error when creating UserIDs table", err)
	}

	//
	// Users
	//
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS Users (
			id INT UNSIGNED AUTO_INCREMENT NOT NULL PRIMARY KEY,
			user_id INT UNSIGNED NOT NULL UNIQUE,
			apple_id VARCHAR(100) NOT NULL,
			name VARCHAR(30) DEFAULT '',
			age TINYINT UNSIGNED DEFAULT 0,
			gender TINYINT DEFAULT 0,
			city_name VARCHAR(30) DEFAULT ''
		);
	`)

	if err != nil {
		log.Println("OpenDatabase, error when creating Users table", err)
	}

	//
	// SessionKeys
	//
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS SessionKeys (
			id INT UNSIGNED AUTO_INCREMENT NOT NULL PRIMARY KEY,
			user_id INT UNSIGNED NOT NULL UNIQUE,
			session_key VARCHAR(100) NOT NULL
		);
	`)

	if err != nil {
		log.Println("OpenDatabase, error when creating SessionKeys table", err)
	}

	//
	// FCMTokens
	//
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS FCMTokens (
			id INT UNSIGNED AUTO_INCREMENT NOT NULL PRIMARY KEY,
			user_id INT UNSIGNED NOT NULL UNIQUE,
			token VARCHAR(512) NOT NULL
		);
	`)

	if err != nil {
		log.Println("OpenDatabase, error when creating FCMTokens table", err)
	}

	return db
}

func GetFCMTokens(userIDs []uint32) ([]string, error) {
	if len(userIDs) == 0 {
		return nil, errors.New("userIDs slice is empty")
	}

	var tokens []string

	args := make([]interface{}, len(userIDs))
	for i, id := range userIDs {
		args[i] = id
	}

	rows, err := MainDB.Query(`
	SELECT token FROM FCMTokens WHERE user_id IN (?`+strings.Repeat(",?", len(args)-1)+`)
		`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var token string
		if err = rows.Scan(&token); err != nil {
			return tokens, err
		}
		tokens = append(tokens, token)
	}

	if err = rows.Err(); err != nil {
		return tokens, err
	}

	return tokens, nil
}

func dsn(dbName string) string {
	return fmt.Sprintf("%s:%s@tcp(%s)/%s", username, password, hostname, dbName)
}

func listDrivers() {
	for _, driver := range sql.Drivers() {
		fmt.Printf("Driver: %v \n", driver)
	}
}
