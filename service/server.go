package service

import (
	log "github.com/asiainfoLDP/datahub/utils/clog"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"os"
)

var SERVICE_PORT string

type mux struct {
}

func (m *mux) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	log.Trace("from", req.RemoteAddr, req.Method, req.URL.RequestURI(), req.Proto)

	http.Error(w, "url not found", http.StatusNotFound)

}

func Server() {

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

	router.GET("/v2/service_plans", ServicePlansGetHandler)

	router.NotFound = &mux{}
	//router.MethodNotAllowed = &mux{}

	log.Info("listening on", SERVICE_PORT)
	err := http.ListenAndServe(SERVICE_PORT, router)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
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

func init() {
	SERVICE_PORT = os.Getenv("SERVICE_PORT")
}
