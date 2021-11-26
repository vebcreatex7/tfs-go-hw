package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/go-chi/chi/v5"
	"github.com/spf13/viper"

	"github.com/sirupsen/logrus"
	"github.com/tfs-go-hw/course_project/config"
	"github.com/tfs-go-hw/course_project/internal/domain"
	"github.com/tfs-go-hw/course_project/internal/handlers"
	"github.com/tfs-go-hw/course_project/internal/repository"
	"github.com/tfs-go-hw/course_project/internal/services"
	"github.com/tfs-go-hw/course_project/internal/services/indicators"
	"github.com/tfs-go-hw/course_project/internal/services/kraken"
	pkglog "github.com/tfs-go-hw/course_project/pkg/log"
	pkgpostgres "github.com/tfs-go-hw/course_project/pkg/postgres"
)

func main() {

	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel)

	// Get private/public key, port, db auth data.
	err := config.Init()
	if err != nil {
		logger.Fatal(err)
	}

	dsn := "postgres://" + viper.GetString("postgres.user") + ":" +
		viper.GetString("postgres.password") + "@localhost:" +
		viper.GetString("postgres.port") + "/" + viper.GetString("postgres.db")

	pool, err := pkgpostgres.NewPool(dsn, logger)
	if err != nil {
		logger.Fatal(err)
	}
	defer pool.Close()

	repo := repository.NewRepository(pool)

	// Kraken services
	kraken := kraken.NewKraken(viper.GetString("keys.public_key"), viper.GetString("keys.private_key"))
	// Indicator services
	macd := indicators.NewMacd()
	// Bot services
	botService := services.NewBotService(repo, logger, kraken, macd)

	r := chi.NewRouter()
	r.Use(pkglog.NewStructuredLogger(logger))

	done, cancelFunc := context.WithCancel(context.Background())

	// REST handler for control a bot
	botHandler := handlers.NewBot(done, botService, logger)
	r.Mount("/bot", botHandler.Routes())
	serv := new(domain.Server)
	go func() {
		if err := serv.Run(":"+viper.GetString("port"), r); err != nil {
			log.Fatalln(err)
		}
	}()

	wg := &sync.WaitGroup{}
	wg.Add(1)

	// Launching the bot
	botHandler.Run(wg)

	sigquit := make(chan os.Signal, 1)
	signal.Notify(sigquit, syscall.SIGINT)
	<-sigquit
	cancelFunc()
	wg.Wait()
	if err := serv.Shutdown(context.Background()); err != nil {
		logger.Println("Can't shutdown main server: ", err.Error())
	}
}
