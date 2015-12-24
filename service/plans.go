package service

import (
	"encoding/json"
	"github.com/asiainfoLDP/datafactory-backingservice-manager/ds"
	log "github.com/asiainfoLDP/datahub/utils/clog"
	"github.com/julienschmidt/httprouter"
	"io/ioutil"
	"net/http"
)

//ok
func ServicePlansGetHandler(rw http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	log.Trace("from", req.RemoteAddr, req.Method, req.URL.RequestURI(), req.Proto)
	//db := getDB()

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Error(err)
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	} else {
		log.Warn(string(body))
	}

	plans, err := dbgetserviceplans()
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	} else {
		log.Debugf("%+v", plans)
		if resp, err := json.Marshal(plans); err != nil {

			log.Error(err)
			http.Error(rw, "server internal error", http.StatusInternalServerError)
			return
		} else {
			rw.Header().Add("X-Content-Type-Options", "nosniff")
			rw.Header().Add("Content-Type", "application/json;charset=utf-8")
			rw.WriteHeader(http.StatusOK)
			rw.Write(resp)
		}
	}

	/*
		catalog := ds.Catalog{}

		log.Infof("%v,%v,%+v", resp.StatusCode, string(body), catalog)
		err = json.Unmarshal(body, &catalog)

		if err != nil {
			log.Error(err)
			http.Error(rw, "server internal error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		log.Infof("%v,%+v", string(body), catalog)

		var count int

		err = db.QueryRow("SELECT COUNT(1) FROM service_brokers WHERE broker_url=? OR name=?", sb.Url, sb.Name).Scan(&count)
		checkSqlErr(err)
		log.Debug("COUNT=", count)

		if count > 0 {
			errStr := sb.Url + " or " + sb.Name + " already exist!"
			http.Error(rw, errStr, http.StatusBadRequest)
			return
		}

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
	*/
	return
}

func dbgetserviceplans() (*ds.ServicePlans, error) {

	db := getDB()

	plans := &ds.ServicePlans{TotalPages: 1}
	/*
		var (
			guid                 string
			name                 string
			description          string
			service_id           string
			extra                string
			unique_id            string
			free, public, active bool
		)
	*/
	resources := []ds.Response{}

	rows, err := db.Query("SELECT guid,created_at,name,description,free,service_id,extra,unique_id,public,active FROM service_plans")
	checkSqlErr(err)
	defer rows.Close()

	for rows.Next() {
		plan := &ds.EntityPlan{}

		var extra, t, service_guid string
		//err := rows.Scan(&guid, &name, &description, &free, &service_id, &extra, &unique_id, &public, &active)
		err := rows.Scan(&plan.Guid, &t, &plan.Name, &plan.Description, &plan.Free,
			&plan.ServiceID, &extra, &plan.UniqueID, &plan.Public, &plan.Active)
		//err := rows.Scan(plan)
		if err != nil {
			log.Fatal(err)
		}

		plans.TotalResult += 1

		err = json.Unmarshal([]byte(extra), &plan.Extra)
		if err != nil {
			log.Error(err)
		}

		err = db.QueryRow("SELECT guid FROM services WHERE id=?", plan.ServiceID).Scan(&service_guid)
		checkSqlErr(err)
		log.Debugf("service_guid %s ", service_guid)
		plan.ServiceID = service_guid
		plan.ServiceUrl = "/v2/services/" + service_guid
		plan.InstanceUrl = "/v2/service_plans/" + plan.Guid + "/service_instances"

		resource := ds.Response{
			Metadata: ds.Metadata{
				Guid:      plan.Guid,
				Create_at: t,
				Url:       "/v2/service_plans/" + plan.Guid,
			},
			Entity: plan,
		}
		resources = append(resources, resource)
		log.Printf("%+v,extra:%v", plan, extra)
	}

	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	plans.Resources = resources

	return plans, err

}
