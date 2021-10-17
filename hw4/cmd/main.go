package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/tfs-go-hw/hw4/internal/domain"
	"github.com/tfs-go-hw/hw4/internal/handler"
)

func main() {
	sigquit := make(chan os.Signal, 1)
	signal.Notify(sigquit, syscall.SIGINT)

	handler := new(handler.Handler)
	ser := new(domain.Server)

	go func() {
		if err := ser.Run(":8000", handler.InitRoutes()); err != nil {
			log.Fatalln(err)
		}
	}()

	<-sigquit

	if err := ser.Shutdown(context.Background()); err != nil {
		log.Println("Can't shutdown main server: ", err.Error())
	}
}
