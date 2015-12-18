package service

import (
	"database/sql"
	"encoding/json"
	"github.com/asiainfoLDP/datafactory-backendservice-manager/ds"
	log "github.com/asiainfoLDP/datahub/utils/clog"
	_ "github.com/go-sql-driver/mysql"
	"github.com/julienschmidt/httprouter"
	uuid "github.com/satori/go.uuid"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

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

	si := ds.BackingServiceInstance{}
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

	siresp := &ds.Response{
		Metadata: ds.Metadata{
			Guid:      guid,                      //"e5ba6b33-f9c6-4f75-ba3f-9bb9a93654df",
			Create_at: t,                         //"2015-11-30T23:39:01Z",
			Url:       req.URL.Path + "/" + guid, //"/v2/service_instances/e5ba6b33-f9c6-4f75-ba3f-9bb9a93654df",
			//Updated_at: nil,
		},
		Entity: ds.EntitySI{
			Name:              si.Name,              // "my-service-instance",
			Service_plan_guid: si.Service_plan_guid, //"b093007b-4c3b-4774-8603-364696a72afa",
			Space_guid:        si.Space_guid,        //"70130716-850c-41a5-a6db-d440fecacc67",
			//Dashboard_url:     nil,
			Type: "managed_service_instance",
			Last_operation: ds.Last_operation{
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

func dbserviceinstance(si *ds.BackingServiceInstance) (guid, t string, err error) {
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
