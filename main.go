package cflogger

import (
	"encoding/json"
	"fmt"
	"github.com/r-franke/cfconfig"
	"github.com/r-franke/cfrabbit/publisher"
	"log"
	"os"
)

type MaintenancePublisher struct {
	publisher *publisher.Publisher
}

type CustomErrorLogger struct {
	*log.Logger
}

//goland:noinspection GoUnusedGlobalVariable
var (
	InfoLogger           *log.Logger
	ErrorLogger          *log.Logger
	MaintenanceLogger    *log.Logger
	internalErrorLogger  *log.Logger
	maintenancePublisher *MaintenancePublisher
	env                  cfconfig.Env
	reportErrors         bool
)

func init() {
	var err error
	env = cfconfig.LoadEnvironment("cflogger-dev", []cfconfig.Request{})

	_, reportErrors = os.LookupEnv("REPORT_ERRORS")

	InfoLogger = log.New(os.Stdout, "", log.Lshortfile)
	InfoLogger.Printf("cflogger: Starting up, REPORT_ERRORS = %t", reportErrors)

	internalErrorLogger = log.New(os.Stderr, "cflogger: ", log.Lshortfile)

	maintenancePublisher = &MaintenancePublisher{}
	pub, err := publisher.NewPublisher("signal.out", "fanout")
	if err != nil {
		internalErrorLogger.Fatalf("cannot get RabbitPublisher\n%s", err.Error())
	}
	maintenancePublisher.publisher = &pub
	MaintenanceLogger = log.New(maintenancePublisher, fmt.Sprintf("%s:\n", env.AppName), log.Lshortfile)

	ErrorLogger = log.New(&CustomErrorLogger{log.New(os.Stderr, "", log.Lshortfile)}, "", 0)
}
func (cel *CustomErrorLogger) Write(p []byte) (n int, err error) {
	if reportErrors {
		MaintenanceLogger.Print(string(p))
	}
	cel.Logger.Print(string(p))
	return len(p), nil
}

func (ml *MaintenancePublisher) Write(p []byte) (n int, err error) {
	signalMsg := Payload{
		Payload: SignalMessage{
			Channel:     "signal",
			Subscribers: []string{"[H] Maintenance"},
			MessageBody: string(p),
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
