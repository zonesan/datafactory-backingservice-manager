package service

import (
	"bytes"
	"database/sql"
	"fmt"
	log "github.com/asiainfoLDP/datahub/utils/clog"
	_ "github.com/go-sql-driver/mysql"
	"net/http"
	"os"
	"strings"
)

var (
	SERVICE_BROKER_API_SERVER string
	SERVICE_BROKER_API_PORT   string
	database                  = new(SqlHandler)
)

type SqlHandler struct {
	db *sql.DB
}

func getDB() *sql.DB {
	return database.db
}

func commToServiceBroker(method, path string, jsonData []byte, header map[string]string) (resp *http.Response, err error) {
	//fmt.Println(method, path, string(jsonData))

	req, err := http.NewRequest(strings.ToUpper(method) /*SERVICE_BROKER_API_SERVER+*/, path, bytes.NewBuffer(jsonData))

	if len(header) > 0 {
		for key, value := range header {
			req.Header.Set(key, value)
		}
	}

	return http.DefaultClient.Do(req)
}

func initDB() {
	DB_ADDR := os.Getenv("MYSQL_ENV_TCP_ADDR")
	DB_PORT := os.Getenv("MYSQL_ENV_TCP_PORT")
	DB_DATABASE := os.Getenv("MYSQL_ENV_DATABASE")
	DB_USER := os.Getenv("MYSQL_ENV_USER")
	DB_PASSWORD := os.Getenv("MYSQL_ENV_PASSWORD")

	//DB_URL := fmt.Sprintf(`%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=true`, DB_USER, DB_PASSWORD, DB_ADDR, DB_PORT, DB_DATABASE)
	DB_URL := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", DB_USER, DB_PASSWORD, DB_ADDR, DB_PORT, DB_DATABASE)

	log.Info("connect to ", DB_URL)
	db, err := sql.Open("mysql", DB_URL)
	if err != nil {
		log.Fatal("error: %s\n", err)
	} else {
		database.db = db
	}

}

/*
`{
  "metadata": {
    "guid": "f2b4df6d-3a7e-44d8-a173-103c21a6ab7e",
    "created_at": "2015-11-30T23:38:45Z",
    "updated_at": null,
    "url": "/v2/service_brokers/f2b4df6d-3a7e-44d8-a173-103c21a6ab7e"
  },
  "entity": {
    "name": "service-broker-name",
    "broker_url": "https://broker.example.com",
    "auth_username": "admin",
    "space_guid": null
  }
}`
*/

func post(url string, jsonStr []byte) (err error) {
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("X-Custom-Header", "myvalue")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	return
}

func checkErr(err error) {
	if err != nil {
		log.Error(err)
	}
}

func insert(db *sql.DB) {
	stmt, err := db.Prepare("INSERT INTO user(username, password) VALUES(?, ?)")
	defer stmt.Close()

	if err != nil {
		log.Println(err)
		return
	}
	stmt.Exec("guotie", "guotie")
	stmt.Exec("testuser", "123123")

}

func init() {
	SERVICE_BROKER_API_SERVER = os.Getenv("SERVICE_BROKER_API_SERVER")
	SERVICE_BROKER_API_PORT = os.Getenv("SERVICE_BROKER_API_PORT")

	SERVICE_BROKER_API_SERVER = "http://" + SERVICE_BROKER_API_SERVER + ":" + SERVICE_BROKER_API_PORT
	log.Info("service broker conn:", SERVICE_BROKER_API_SERVER)

	initDB()
}
