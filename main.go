package main

import (
	"encoding/json"
	"fmt"
	"github.com/r-franke/cfconfig"
	"github.com/r-franke/cfrabbit"
	"log"
	"os"
)

type MaintenancePublisher struct {
	publisher *cfrabbit.Publisher
}

//goland:noinspection GoUnusedGlobalVariable
var (
	InfoLogger           *log.Logger
	ErrorLogger          *log.Logger
	MaintenanceLogger    *log.Logger
	internalErrorLogger  *log.Logger
	maintenancePublisher *MaintenancePublisher
	env                  cfconfig.Env
)

func init() {
	var err error
	env = cfconfig.LoadEnvironment("cflogger-dev", []cfconfig.Request{})
	InfoLogger = log.New(os.Stdout, "", log.Lshortfile)
	ErrorLogger = log.New(os.Stderr, "", log.Lshortfile)
	internalErrorLogger = log.New(os.Stderr, "cflogger: ", log.Lshortfile)

	maintenancePublisher.publisher, err = cfrabbit.NewPublisher("signal.out", "fanout")
	if err != nil {
		internalErrorLogger.Fatalf("cannot get RabbitPublisher\n%s", err.Error())
	}
	MaintenanceLogger = log.New(maintenancePublisher, env.AppName, log.Lshortfile)
}

func (ml *MaintenancePublisher) Write(p []byte) (n int, err error) {
	signalMsg := Payload{
		Payload: SignalMessage{
			Channel:     "signal",
			Subscribers: []string{"[H] Maintenance"},
			MessageBody: fmt.Sprintf("%s: %s", env.AppName, string(p)),
		},
	}

	marshalledMsg, err := json.Marshal(signalMsg)
	if err != nil {
		internalErrorLogger.Printf("error marshalling message: %v\n%s\n", signalMsg, err.Error())
		return -1, err
	}
	err = ml.publisher.Publish("", marshalledMsg)
	if err != nil {
		internalErrorLogger.Printf("Could not send maintenance message: %s", string(p))
		return -1, err
	}
	return len(p), nil
}
