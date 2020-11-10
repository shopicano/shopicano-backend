package cmd

import (
	"github.com/shopicano/shopicano-backend/app"
	"github.com/shopicano/shopicano-backend/config"
	"github.com/shopicano/shopicano-backend/log"
	"github.com/shopicano/shopicano-backend/machinery"
	payment_gateways "github.com/shopicano/shopicano-backend/payment-gateways"
	"github.com/spf13/cobra"
	"os"
)

var workerCmd = &cobra.Command{
	Use:    "worker",
	Short:  "Worker starts workers for async task processing",
	PreRun: preServeWorker,
	Run:    serveWorker,
}

func preServeWorker(cmd *cobra.Command, args []string) {
	if err := app.ConnectMinio(); err != nil {
		log.Log().Errorln("Failed to connect to minio : ", err)
		os.Exit(-1)
	}
	if err := machinery.NewRabbitMQConnection(); err != nil {
		log.Log().Errorln("Failed to connect to rabbitmq : ", err)
		os.Exit(-1)
	}
	if err := machinery.RegisterRabbitMQTasks(); err != nil {
		log.Log().Errorln("Failed to register rabbitmq tasks : ", err)
		os.Exit(-1)
	}
}

func serveWorker(cmd *cobra.Command, args []string) {
	if err := payment_gateways.SetActivePaymentGateway(config.PaymentGateway()); err != nil {
		log.Log().Errorln("Failed to setup payment gateway : ", err)
		os.Exit(-1)
	}

	machinery.RunRabbitMQWorker()
}
