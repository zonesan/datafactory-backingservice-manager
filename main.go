package main

import (
	"github.com/asiainfoLDP/datafactory-backingservice-manager/service"
	log "github.com/asiainfoLDP/datahub/utils/clog"
)

func init() {
	log.SetLogLevel(log.LOG_LEVEL_DEBUG)
}

func main() {
	service.Server()
}
