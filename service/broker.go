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

	sb := ds.BackingServiceBroker{}
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

	header := make(map[string]string)
	header["Content-Type"] = "application/json"
	header["Authorization"] = basicAuthStr(sb.AuthName, sb.AuthPass)

	resp, err := commToServiceBroker("GET", sb.Url+"/v2/catalog", nil, header)
	if err != nil {
		log.Error(err)
		http.Error(rw, "server internal error:"+err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error(err)
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	catalog := ds.Catalog{}

	log.Infof("%v,%v,%+v", resp.StatusCode, string(body), catalog)
	err = json.Unmarshal(body, &catalog)

	if err != nil {
		log.Error(err)
		http.Error(rw, "server internal error: "+err.Error(), http.StatusInternalServerError)
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

	sbresp := &ds.Response{
		Metadata: ds.Metadata{
			Guid:      guid,                      //"f2b4df6d-3a7e-44d8-a173-103c21a6ab7e",
			Create_at: t,                         //"2015-11-30T23:38:45Z",
			Url:       req.URL.Path + "/" + guid, // "/v2/service_brokers/f2b4df6d-3a7e-44d8-a173-103c21a6ab7e",
			//Updated_at: string(nil),
		},
		Entity: ds.EntitySB{
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

func dbservicebroker(catalog ds.Catalog, service_broker_id int64) (err error) {
	db := getDB()

	for _, s := range catalog.Services {
		if r, err := db.Exec(`INSERT INTO services
			(guid,label,provider,description,bindable,plan_updateable,service_broker_id,purging)
			VALUES(?,?,?,?,?,?,?,?)`,
			s.Id, s.Name, s.Name, s.Description, s.Bindable, s.PlanUpdateable, service_broker_id, 0); err != nil {
			log.Error("INSERT INTO services error:", err)
		} else {
			service_id, err := r.LastInsertId()
			for _, plan := range s.Plans {
				if _, err = db.Exec(`INSERT INTO service_plans
						(guid,name,description,free,service_id,unique_id)
						VALUES(?,?,?,?,?,?)`,
					plan.Id, plan.Name, plan.Description, plan.Free, service_id, plan.Id); err != nil {
					log.Error("INSERT INTO service_plans", err)
				}
			}
		}

		if _, err = db.Exec(`INSERT INTO service_dashboard_clients
			(uaa_id,service_broker_id) VALUES(?,?)`,
			s.DashboardClient.Id, service_broker_id); err != nil {
			log.Error("INSERT INTO service_dashboard_clients error:", err)

		}

	}
	return
}
