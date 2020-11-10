package cmd

import (
	"github.com/shopicano/shopicano-backend/app"
	"github.com/shopicano/shopicano-backend/config"
	"github.com/shopicano/shopicano-backend/log"
	"github.com/shopicano/shopicano-backend/machinery"
	payment_gateways "github.com/shopicano/shopicano-backend/payment-gateways"
	"github.com/shopicano/shopicano-backend/server"
	"github.com/spf13/cobra"
	"os"
)

var serveCmd = &cobra.Command{
	Use:    "serve",
	Short:  "Serve starts http server",
	PreRun: preServe,
	Run:    serve,
}

func preServe(cmd *cobra.Command, args []string) {
	if err := app.ConnectMinio(); err != nil {
		log.Log().Errorln("Failed to connect to minio : ", err)
		os.Exit(-1)
	}

	if err := machinery.NewRabbitMQConnection(); err != nil {
		log.Log().Errorln("Failed to connect to rabbitmq : ", err)
		os.Exit(-1)
	}
}

func serve(cmd *cobra.Command, args []string) {
	if err := payment_gateways.SetActivePaymentGateway(config.PaymentGateway()); err != nil {
		log.Log().Errorln("Failed to setup payment gateway : ", err)
		os.Exit(-1)
	}
	server.StartServer()
}
