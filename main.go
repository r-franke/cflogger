package main

import (
	"encoding/json"
	"fmt"
	"github.com/r-franke/cfconfig"
	"github.com/r-franke/cfrabbit"
	"log"
	"os"
)

type MaintenanceLog struct {
	publisher *cfrabbit.Publisher
}

//goland:noinspection GoUnusedGlobalVariable
var (
	MaintenanceLogger   *MaintenanceLog
	InfoLogger          *log.Logger
	ErrorLogger         *log.Logger
	internalErrorLogger *log.Logger
	env                 cfconfig.Env
)

func init() {
	var err error
	env = cfconfig.LoadEnvironment("cflogger-dev", []cfconfig.Request{})
	InfoLogger = log.New(os.Stdout, "", log.Lshortfile)
	ErrorLogger = log.New(os.Stderr, "", log.Lshortfile)
	internalErrorLogger = log.New(os.Stderr, "cflogger: ", log.Lshortfile)

	MaintenanceLogger.publisher, err = cfrabbit.NewPublisher("signal.out", "fanout")
	if err != nil {
		internalErrorLogger.Fatalf("cannot get RabbitPublisher\n%s", err.Error())
	}
}

func (ml *MaintenanceLog) Printf(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	signalMsg := Payload{
		Payload: SignalMessage{
			Channel:     "signal",
			Subscribers: []string{"[H] Maintenance"},
			MessageBody: fmt.Sprintf("%s: %s", env.AppName, msg),
		},
	}

	marshalledMsg, err := json.Marshal(signalMsg)
	if err != nil {
		internalErrorLogger.Printf("error marshalling message: %v\n%s\n", signalMsg, err.Error())
	}
	err = ml.publisher.Publish("", marshalledMsg)
	if err != nil {
		internalErrorLogger.Printf("Could not send maintenance message: %s", msg)
	}
}

func main() {

}
