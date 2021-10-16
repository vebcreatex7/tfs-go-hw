package main

import (
	"log"

	"github.com/tfs-go-hw/hw4/internal/domain"
	"github.com/tfs-go-hw/hw4/internal/handler"
)

func main() {

	handler := new(handler.Handler)
	ser := new(domain.Server)
	if err := ser.Run("5000", handler.InitRoutes()); err != nil {
		log.Fatalln(err)
	}
}
