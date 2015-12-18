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
	"time"
)

func ServiceBindingsPostHandler(rw http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	log.Trace("from", req.RemoteAddr, req.Method, req.URL.RequestURI(), req.Proto)

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Error(err)
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	sbind := ds.BackingServiceBinding{}
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

	sbindresp := &ds.Response{
		Metadata: ds.Metadata{
			Guid:      guid,                      //"f2b4df6d-3a7e-44d8-a173-103c21a6ab7e",
			Create_at: t,                         //"2015-11-30T23:38:45Z",
			Url:       req.URL.Path + "/" + guid, //"/v2/service_brokers/f2b4df6d-3a7e-44d8-a173-103c21a6ab7e",
			//Updated_at: string(nil),
		},
		Entity: ds.EntitySBind{
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

func dbservicebinding(sb *ds.BackingServiceBinding) (guid, t string, err error) {
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
		if _, err = db.Exec(`INSERT INTO service_bindings
			(guid,created_at,service_instance_id,app_id) VALUES(?,?,?,?)`,
			guid, t, service_instance_id, sb.App_guid); err != nil {
			log.Error("INSERT INTO service_bindings error:", err)

		}
	}

	return
}
