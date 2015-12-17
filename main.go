package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	log "github.com/asiainfoLDP/datahub/utils/clog"
	_ "github.com/go-sql-driver/mysql"
	"github.com/julienschmidt/httprouter"
	uuid "github.com/satori/go.uuid"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

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

var (
	SERVICE_PORT              string
	SERVICE_BROKER_API_SERVER string
	ds                        = new(Ds)
)

type Ds struct {
	db *sql.DB
}

func rootHandler(rw http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	log.Trace("from", req.RemoteAddr, req.Method, req.URL.RequestURI(), req.Proto)
	/*
		for k, v := range req.Header {
			fmt.Printf("[%s]=[%s]\n", k, v)
		}
	*/
	http.Error(rw, req.URL.Path, http.StatusForbidden)
}

//ok
func ServiceBrokerPostHandler(rw http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	log.Trace("from", req.RemoteAddr, req.Method, req.URL.RequestURI(), req.Proto)
	db := getDB()
	guid := uuid.NewV4().String()
	t := time.Now().Format("2006-01-02T15:04:05") //time.Now().Format(time.RFC3339)

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Error(err)
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	sb := BackingServiceBroker{}
	if err = json.Unmarshal(body, &sb); err != nil {
		log.Error(err)
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
	log.Debugf("%v", sb)

	if len(sb.AuthName) == 0 || len(sb.AuthPass) == 0 ||
		len(sb.Name) == 0 || len(sb.Url) == 0 {
		http.Error(rw, "invalid argument.", http.StatusBadRequest)
		return
	}

	resp, err := commToServiceBroker("GET", "/v2/catalog", nil, nil)
	if err != nil {
		log.Error(err)
		http.Error(rw, "server internal error", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error(err)
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	catalog := Catalog{}

	err = json.Unmarshal(body, &catalog)

	if err != nil {
		log.Error(err)
		http.Error(rw, "server internal error"+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Infof("%v,%+v", string(body), catalog)

	if r, err := db.Exec(`INSERT INTO service_brokers(guid,created_at,name,broker_url,auth_password,auth_username)
			VALUES(?,?,?,?,?,?)`, guid, t, sb.Name, sb.Url, sb.AuthPass, sb.AuthName); err != nil {
		log.Error("INSERT INTO service_brokers", err)
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	} else {
		service_broker_id, _ := r.LastInsertId()
		dbservicebroker(catalog, service_broker_id)
	}

	sbresp := &Response{
		Metadata: Metadata{
			Guid:      guid,                      //"f2b4df6d-3a7e-44d8-a173-103c21a6ab7e",
			Create_at: t,                         //"2015-11-30T23:38:45Z",
			Url:       req.URL.Path + "/" + guid, // "/v2/service_brokers/f2b4df6d-3a7e-44d8-a173-103c21a6ab7e",
			//Updated_at: string(nil),
		},
		Entity: EntitySB{
			Name:          sb.Name,     //  "service-broker-name",
			Broker_url:    sb.Url,      //"https://broker.example.com",
			Auth_username: sb.AuthName, //"admin",
			//Space_guid:    guid,
		},
	}

	if resp, err := json.Marshal(sbresp); err != nil {
		http.Error(rw, "server internal error", http.StatusInternalServerError)
	} else {
		rw.Header().Add("X-Content-Type-Options", "nosniff")
		rw.Header().Add("X-VCAP-Request-ID", "")
		rw.Header().Add("Content-Type", "application/json;charset=utf-8")
		rw.WriteHeader(http.StatusCreated)
		rw.Write(resp)
	}

	return
}

func dbservicebroker(catalog Catalog, service_broker_id int64) (err error) {
	db := getDB()

	for _, s := range catalog.Services {
		if r, err := db.Exec(`INSERT INTO services
			(guid,label,provider,description,bindable,plan_updateable,purging)
			VALUES(?,?,?,?,?,?,?)`,
			s.Id, s.Name, s.Name, s.Description, s.Bindable, s.PlanUpdateable, 0); err != nil {
			log.Error("INSERT INTO services error:", err)
		} else {
			service_id, err := r.LastInsertId()
			for _, plan := range s.Plans {
				if _, err = db.Exec(`INSERT INTO service_plans
						(guid,name,description,free,service_id)
						VALUES(?,?,?,?,?)`,
					plan.Id, plan.Name, plan.Description, plan.Free, service_id); err != nil {
					log.Error("INSERT INTO service_plans", err)
				}
			}
		}

		if _, err := db.Exec(`INSERT INTO service_dashboard_clients
			(uaa_id,service_broker_id) VALUES(?,?)`,
			s.DashboardClient.Id, service_broker_id); err != nil {
			log.Error("INSERT INTO service_dashboard_clients error:", err)

		}

	}
	return
}

func ServiceInstancesPostHandler(rw http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	log.Trace("from", req.RemoteAddr, req.Method, req.URL.RequestURI(), req.Proto)

	accepts_incomplete, err := strconv.ParseBool(ps.ByName("accepts_incomplete"))
	if err == nil && accepts_incomplete == true {

	}

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Error(err)
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	si := BackingServiceInstance{}
	if err = json.Unmarshal(body, &si); err != nil {
		log.Error(err)
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
	log.Debugf("%v", si)

	if len(si.Name) == 0 || len(si.Service_plan_guid) == 0 || len(si.Space_guid) == 0 {
		http.Error(rw, "invalid argument.", http.StatusBadRequest)
		return
	}

	guid, t, err := dbserviceinstance(&si)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	siresp := &Response{
		Metadata: Metadata{
			Guid:      guid,                      //"e5ba6b33-f9c6-4f75-ba3f-9bb9a93654df",
			Create_at: t,                         //"2015-11-30T23:39:01Z",
			Url:       req.URL.Path + "/" + guid, //"/v2/service_instances/e5ba6b33-f9c6-4f75-ba3f-9bb9a93654df",
			//Updated_at: nil,
		},
		Entity: EntitySI{
			Name:              si.Name,              // "my-service-instance",
			Service_plan_guid: si.Service_plan_guid, //"b093007b-4c3b-4774-8603-364696a72afa",
			Space_guid:        si.Space_guid,        //"70130716-850c-41a5-a6db-d440fecacc67",
			//Dashboard_url:     nil,
			Type: "managed_service_instance",
			Last_operation: Last_operation{
				Type:  "create",
				State: "in progress",
				//Updated_at: nil,
				Created_at: "2015-11-30T23:39:01Z",
			},
			Space_url:           "/v2/spaces/" + si.Space_guid,               //70130716-850c-41a5-a6db-d440fecacc67",
			Service_plan_url:    "/v2/service_plans/" + si.Service_plan_guid, //b093007b-4c3b-4774-8603-364696a72afa",
			Service_binding_url: "/v2/service_instances/" + guid,             //e5ba6b33-f9c6-4f75-ba3f-9bb9a93654df/service_bindings",
			Routes_url:          "/v2/service_instances/" + guid,             //e5ba6b33-f9c6-4f75-ba3f-9bb9a93654df/routes",
			Tags:                si.Tags,
		},
	}

	if resp, err := json.Marshal(siresp); err != nil {
		http.Error(rw, "server internal error", http.StatusInternalServerError)
	} else {
		rw.Header().Add("X-Content-Type-Options", "nosniff")
		rw.Header().Add("Content-Type", "application/json;charset=utf-8")
		rw.WriteHeader(http.StatusAccepted)
		rw.Write(resp)
	}

	return
}

func dbserviceinstance(si *BackingServiceInstance) (guid, t string, err error) {
	db := getDB()
	guid = uuid.NewV4().String()
	t = time.Now().Format("2006-01-02T15:04:05") //time.Now().Format(time.RFC3339)

	var service_plan_id string
	err = db.QueryRow("SELECT id FROM service_plans WHERE guid=?", si.Service_plan_guid).Scan(&service_plan_id)
	switch {
	case err == sql.ErrNoRows:
		log.Error("No service_plan_id with that Service_plan_guid.", err)
	case err != nil:
		log.Fatal(err)
	default:
		if _, err = db.Exec(`INSERT INTO service_instances
			(guid,created_at,name,service_plan_id,tags) VALUES(?,?,?,?,?)`,
			guid, t, si.Name, service_plan_id, strings.Join(si.Tags, ", ")); err != nil {
			log.Error("INSERT INTO service_instances error:", err)

		}
	}

	return

}

func ServiceBindingsPostHandler(rw http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	log.Trace("from", req.RemoteAddr, req.Method, req.URL.RequestURI(), req.Proto)

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Error(err)
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	sbind := BackingServiceBinding{}
	if err = json.Unmarshal(body, &sbind); err != nil {
		log.Error(err)
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
	log.Debugf("%v", sbind)

	if len(sbind.Service_instance_guid) == 0 || len(sbind.App_guid) == 0 {
		http.Error(rw, "invalid argument.", http.StatusBadRequest)
		return
	}

	guid, t, err := dbservicebinding(&sbind)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	sbindresp := &Response{
		Metadata: Metadata{
			Guid:      guid,                      //"f2b4df6d-3a7e-44d8-a173-103c21a6ab7e",
			Create_at: t,                         //"2015-11-30T23:38:45Z",
			Url:       req.URL.Path + "/" + guid, //"/v2/service_brokers/f2b4df6d-3a7e-44d8-a173-103c21a6ab7e",
			//Updated_at: string(nil),
		},
		Entity: EntitySBind{
			App_guid:              sbind.App_guid,
			Service_instance_guid: sbind.Service_instance_guid,
			App_url:               "/v2/apps/" + sbind.App_guid,
			Service_instance_url:  "/v2/user_provided_service_instances/" + sbind.Service_instance_guid,
		},
	}

	if resp, err := json.Marshal(sbindresp); err != nil {
		http.Error(rw, "server internal error", http.StatusInternalServerError)
	} else {
		rw.Header().Add("X-Content-Type-Options", "nosniff")
		rw.Header().Add("Content-Type", "application/json;charset=utf-8")
		rw.WriteHeader(http.StatusCreated)
		rw.Write(resp)
	}

	return
}

func dbservicebinding(sb *BackingServiceBinding) (guid, t string, err error) {
	db := getDB()
	guid = uuid.NewV4().String()
	t = time.Now().Format("2006-01-02T15:04:05") //time.Now().Format(time.RFC3339)

	var service_instance_id string
	err = db.QueryRow("SELECT id FROM service_instances WHERE guid=?", sb.Service_instance_guid).Scan(&service_instance_id)

	switch {
	case err == sql.ErrNoRows:
		log.Error("No service_instance_id with that Service_instance_guid.", err)
	case err != nil:
		log.Fatal(err)
	default:
		if _, err := db.Exec(`INSERT INTO service_bindings
			(guid,created_at,service_instance_id,app_id) VALUES(?,?,?,?)`,
			guid, t, service_instance_id, sb.App_guid); err != nil {
			log.Error("INSERT INTO service_bindings error:", err)

		}
	}

	return
}

func main() {

	if len(SERVICE_PORT) == 0 {
		SERVICE_PORT = "41000"
		log.Info("no $SERVICE_PORT found, use default", SERVICE_PORT)
	}
	SERVICE_PORT = ":" + SERVICE_PORT

	router := httprouter.New()

	router.GET("/", rootHandler)
	//router.GET("/debug", rootHandler)

	router.POST("/v2/service_brokers", ServiceBrokerPostHandler)
	router.POST("/v2/service_instances", ServiceInstancesPostHandler)
	router.POST("/v2/service_bindings", ServiceBindingsPostHandler)

	router.NotFound = &mux{}
	//router.MethodNotAllowed = &mux{}

	log.Info("listening on", SERVICE_PORT)
	err := http.ListenAndServe(SERVICE_PORT, router)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

type mux struct {
}

func (m *mux) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	log.Trace("from", req.RemoteAddr, req.Method, req.URL.RequestURI(), req.Proto)

	http.Error(w, "url not found", http.StatusNotFound)

}

func init() {

	SERVICE_BROKER_API_SERVER = os.Getenv("SERVICE_BROKER_API_SERVER")

	SERVICE_BROKER_API_SERVER = "http://localhost:8001"

	SERVICE_PORT = os.Getenv("SERVICE_PORT")

	log.SetLogLevel(log.LOG_LEVEL_DEBUG)

	initDB()

}

func initDB() {
	DB_ADDR := os.Getenv("MYSQL_PORT_3306_TCP_ADDR")
	DB_PORT := os.Getenv("MYSQL_PORT_3306_TCP_PORT")
	DB_DATABASE := os.Getenv("MYSQL_ENV_MYSQL_DATABASE")
	DB_USER := os.Getenv("MYSQL_ENV_MYSQL_USER")
	DB_PASSWORD := os.Getenv("MYSQL_ENV_MYSQL_PASSWORD")

	DB_ADDR = "127.0.0.1"
	DB_PORT = "3306"
	DB_USER = "root"
	DB_PASSWORD = "zxcvbnm"
	DB_DATABASE = "datafactory"

	//DB_URL := fmt.Sprintf(`%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=true`, DB_USER, DB_PASSWORD, DB_ADDR, DB_PORT, DB_DATABASE)
	DB_URL := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", DB_USER, DB_PASSWORD, DB_ADDR, DB_PORT, DB_DATABASE)

	log.Info("connect to ", DB_URL)
	db, err := sql.Open("mysql", DB_URL)
	if err != nil {
		log.Fatal("error: %s\n", err)
	} else {
		ds.db = db
	}

}

func getDB() *sql.DB {
	return ds.db
}

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

func commToServiceBroker(method, path string, jsonData []byte, header map[string]string) (resp *http.Response, err error) {
	//fmt.Println(method, path, string(jsonData))

	req, err := http.NewRequest(strings.ToUpper(method), SERVICE_BROKER_API_SERVER+path, bytes.NewBuffer(jsonData))

	if len(header) > 0 {
		for key, value := range header {
			req.Header.Set(key, value)
		}
	}

	return http.DefaultClient.Do(req)
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
