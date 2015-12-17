package main

type Metadata struct {
	Guid       string      `json:"guid"`
	Create_at  string      `json:"created_at"`
	Updated_at interface{} `json:"updated_at"`
	Url        string      `json:"url"`
}

type BackingServiceBroker struct {
	Name     string `json:"name,omitempty"`
	Url      string `json:"broker_url,omitempty"`
	AuthName string `json:"auth_username,omitempty"`
	AuthPass string `json:"auth_password,omitempty"`
}

type EntitySB struct {
	Name          string      `json:"name"`
	Broker_url    string      `json:"broker_url"`
	Auth_username string      `json:"auth_username"`
	Space_guid    interface{} `json:"space_guid"`
}

type Response struct {
	Metadata Metadata    `json:"metadata,omitempty"`
	Entity   interface{} `json:"entity,omitempty"`
}

type Parameters struct {
	The_service_broker string `json:"the_service_broker"`
}

type BackingServiceInstance struct {
	Name              string     `json:"name"`
	Service_plan_guid string     `json:"service_plan_guid"`
	Space_guid        string     `json:"space_guid"`
	Parameters        Parameters `json:"parameters,omitempty"`
	Tags              []string   `json:"tags,omitempty"`
}

type Credentials struct{}

type Last_operation struct {
	Type        string `json:"type"`
	State       string `json:"state"`
	Description string `json:"description"`
	Updated_at  string `json:"updated_at"`
	Created_at  string `json:"created_at"`
}

type EntitySI struct {
	Name                string         `json:"name"`
	Credentials         Credentials    `json:"credentials"`
	Service_plan_guid   string         `json:"service_plan_guid"`
	Space_guid          string         `json:"space_guid"`
	Dashboard_url       string         `json:"dashboard_url"`
	Type                string         `json:"type"`
	Last_operation      Last_operation `json:"last_operation"`
	Space_url           string         `json:"space_url"`
	Service_plan_url    string         `json:"service_plan_url"`
	Service_binding_url string         `json:"service_binding_url"`
	Routes_url          string         `json:"routes_url"`
	Tags                []string       `json:"tags"`
}

type BackingServiceBinding struct {
	Service_instance_guid string     `json:"service_instance_guid"`
	App_guid              string     `json:"app_guid"`
	Parameters            Parameters `json:"parameters,omitempty"`
}

type EntitySBind struct {
	Service_instance_guid string      `json:"service_instance_guid"`
	App_guid              string      `json:"app_guid"`
	Credentials           interface{} `json:"credentials"`
	Binding_options       interface{} `json:"binding_options"`
	Gateway_data          interface{} `json:"gateway_data"`
	Gateway_name          string      `json:"gateway_name"`
	Syslog_drain_url      interface{} `json:"syslog_drain_url"`
	App_url               string      `json:"app_url"`
	Service_instance_url  string      `json:"service_instance_url"`
}

/*
   "app_guid": "72ae0608-e822-4660-a97a-ab70cf9017a7",
   "service_instance_guid": "2909e1b9-1e70-42e6-a6e1-67d2fa81ee71",
   "credentials": {
     "creds-key-390": "creds-val-390"
   },
   "binding_options": {

   },
   "gateway_data": null,
   "gateway_name": "",
   "syslog_drain_url": null,
   "app_url": "/v2/apps/72ae0608-e822-4660-a97a-ab70cf9017a7",
   "service_instance_url": "/v2/user_provided_service_instances/2909e1b9-1e70-42e6-a6e1-67d2fa81ee71"

*/

/*
service broker datastruct below
*/

type Catalog struct {
	Services []Service `json:"services"`
}

type ServiceBinding struct {
	Id                string `json:"id"`
	ServiceId         string `json:"service_id"`
	AppId             string `json:"app_id"`
	ServicePlanId     string `json:"service_plan_id"`
	PrivateKey        string `json:"private_key"`
	ServiceInstanceId string `json:"service_instance_id"`
}

type CreateServiceBindingResponse struct {
	// SyslogDrainUrl string      `json:"syslog_drain_url, omitempty"`
	Credentials interface{} `json:"credentials"`
}

type Credential struct {
	Uri      string `json:uri`
	Username string `json:username`
	Password string `json:password`
	Host     string `json:host`
	Port     int    `json:port`
	Database string `json:database`
}

type DashboardClient struct {
	Id          string `json:"id"`
	Secret      string `json:"secret"`
	RedirectUrl string `json:"redirect_uri"`
}
type Service struct {
	Name           string   `json:"name"`
	Id             string   `json:"id"`
	Description    string   `json:"description"`
	Bindable       bool     `json:"bindable"`
	PlanUpdateable bool     `json:"plan_updateable, omitempty"`
	Tags           []string `json:"tags, omitempty"`
	Requires       []string `json:"requires, omitempty"`

	Metadata        interface{}     `json:"metadata, omitempty"`
	Plans           []ServicePlan   `json:"plans"`
	DashboardClient DashboardClient `json:"dashboard_client"`
}

type ServiceInstance struct {
	Id               string `json:"id"`
	DashboardUrl     string `json:"dashboard_url"`
	InternalId       string `json:"internalId, omitempty"`
	ServiceId        string `json:"service_id"`
	PlanId           string `json:"plan_id"`
	OrganizationGuid string `json:"organization_guid"`
	SpaceGuid        string `json:"space_guid"`

	LastOperation *LastOperation `json:"last_operation, omitempty"`

	Parameters interface{} `json:"parameters, omitempty"`
}

type LastOperation struct {
	State                    string `json:"state"`
	Description              string `json:"description"`
	AsyncPollIntervalSeconds int    `json:"async_poll_interval_seconds, omitempty"`
}

type CreateServiceInstanceResponse struct {
	DashboardUrl  string         `json:"dashboard_url"`
	LastOperation *LastOperation `json:"last_operation, omitempty"`
}

type ServicePlan struct {
	Name        string      `json:"name"`
	Id          string      `json:"id"`
	Description string      `json:"description"`
	Metadata    interface{} `json:"metadata, omitempty"`
	Free        bool        `json:"free, omitempty"`
}
