package service

import (
	"encoding/json"
	"github.com/asiainfoLDP/datafactory-backingservice-manager/ds"
	log "github.com/asiainfoLDP/datahub/utils/clog"
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

	guid, t, cred, err := dbservicebinding(&sbind)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
	log.Warn(cred)

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
			Credentials:           cred,
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

func dbservicebinding(sb *ds.BackingServiceBinding) (guid, t string, cred /*ds.Credential*/ interface{}, err error) {
	db := getDB()
	guid = uuid.NewV4().String()
	t = time.Now().Format("2006-01-02T15:04:05") //time.Now().Format(time.RFC3339)

	var (
		service_instance_id string
		service_plan_id     string
		service_plan_guid   string
		service_id          string
		service_guid        string
		service_broker_id   string
		broker_url          string
		username            string
		password            string
	)
	err = db.QueryRow("SELECT id, service_plan_id FROM service_instances WHERE guid=?", sb.Service_instance_guid).Scan(&service_instance_id, &service_plan_id)
	checkSqlErr(err)
	log.Debugf("service_instance_id %s service_plan_id %s", service_instance_id, service_plan_id)

	err = db.QueryRow("SELECT service_id, guid FROM service_plans WHERE id=?", service_plan_id).Scan(&service_id, &service_plan_guid)
	checkSqlErr(err)
	log.Debugf("service_id %s service_plan_guid %s", service_id, service_plan_guid)

	err = db.QueryRow("SELECT guid ,service_broker_id FROM services WHERE id=?", service_id).Scan(&service_guid, &service_broker_id)
	checkSqlErr(err)
	log.Debugf("service_guid %s", service_guid)

	err = db.QueryRow("SELECT broker_url,auth_username,auth_password FROM service_brokers WHERE id=?", service_broker_id).Scan(&broker_url, &username, &password)
	checkSqlErr(err)
	log.Debugf("broker_url %s  username %s password %s", broker_url, username, password)

	binding := &ds.ServiceBinding{
		ServiceId:     service_guid,
		AppId:         sb.App_guid,
		ServicePlanId: service_plan_guid,
	}

	jsonData, err := json.Marshal(binding)
	if err != nil {
		return guid, t, cred, err
	}

	header := make(map[string]string)
	header["Content-Type"] = "application/json"
	header["Authorization"] = basicAuthStr(username, password)

	resp, err := commToServiceBroker("PUT", broker_url+"/v2/service_instances/"+sb.Service_instance_guid+"/service_bindings/"+guid, jsonData, header)
	if err != nil {
		log.Error(err)
		return guid, t, cred, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error(err)
		return guid, t, cred, err
	}

	//cred = &ds.Credential{}

	if len(body) > 0 {
		err = json.Unmarshal(body, &cred)

		if err != nil {
			log.Error(err)
			return guid, t, cred, err
		}
	}

	log.Debugf("%+v, %+v", string(body), cred)

	if _, err = db.Exec(`INSERT INTO service_bindings
			(guid,created_at,service_instance_id,app_guid,credentials) VALUES(?,?,?,?,?)`,
		guid, t, service_instance_id, sb.App_guid, string(body)); err != nil {
		/* TODO APP_ID MUST BE SELECT FROM APP TABLE.sb.App_guid */
		log.Error("INSERT INTO service_bindings error:", err)

	}
	log.Warn("/* FIXED? APP_ID INSTEADED BY APP_GUID, IT SHOULD SELECT FROM APP TABLE.  */")
	return guid, t, cred, nil
}

/*
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
*/
