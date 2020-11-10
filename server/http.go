package server

import (
	"context"
	"fmt"
	"github.com/shopicano/shopicano-backend/config"
	"github.com/shopicano/shopicano-backend/log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func StartServer() {
	addr := fmt.Sprintf("%s:%d", config.App().Base, config.App().Port)

	stop := make(chan os.Signal)
	signal.Notify(stop, os.Interrupt)

	hServer := http.Server{
		Addr:    addr,
		Handler: GetRouter(),
	}

	go func() {
		log.Log().Infoln("Http server has been started on", addr)
		if err := hServer.ListenAndServe(); err != nil {
			log.Log().Errorln("Failed to start http server on :", err)
			os.Exit(-1)
		}
	}()

	<-stop

	ctx, _ := context.WithTimeout(context.Background(), time.Second*5)
	if err := hServer.Shutdown(ctx); err != nil {
		log.Log().Infoln("Http server couldn't shutdown gracefully")
	}
	log.Log().Infoln("Http server has been shutdown gracefully")
}
