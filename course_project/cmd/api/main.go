package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/go-chi/chi/v5"
	"github.com/spf13/viper"
	"github.com/tfs-go-hw/course_project/config"
	"github.com/tfs-go-hw/course_project/internal/domain"
	"github.com/tfs-go-hw/course_project/internal/handlers"
	"github.com/tfs-go-hw/course_project/internal/services"
	"github.com/tfs-go-hw/course_project/internal/services/indicators"
	"github.com/tfs-go-hw/course_project/internal/services/kraken"
)

func main() {

	// Set private/public key, port.
	err := config.Init()
	if err != nil {
		log.Fatalln(err)
	}

	done, cancelFunc := context.WithCancel(context.Background())
	sigquit := make(chan os.Signal, 1)
	signal.Notify(sigquit, syscall.SIGINT)
	wg := &sync.WaitGroup{}
	wg.Add(1)

	r := chi.NewRouter()

	kraken := kraken.NewKraken(viper.GetString("keys.public_key"), viper.GetString("keys.private_key"))
	macd := indicators.NewMacd()
	botService := services.NewBotService(nil, kraken, macd)

	botHandler := handlers.NewBot(done, botService)

	r.Mount("/bot", botHandler.Routes())

	serv := new(domain.Server)

	go func() {
		if err := serv.Run(":"+viper.GetString("port"), r); err != nil {
			log.Fatalln(err)
		}
	}()

	botHandler.Run(wg)

	<-sigquit
	cancelFunc()
	fmt.Println("Cancel")
	wg.Wait()
	if err := serv.Shutdown(context.Background()); err != nil {
		log.Println("Can't shutdown main server: ", err.Error())
	}

}
