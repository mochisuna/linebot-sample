package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/mochisuna/linebot-sample/application"
	"github.com/mochisuna/linebot-sample/config"
	"github.com/mochisuna/linebot-sample/handler"
	"github.com/mochisuna/linebot-sample/infrastructure"
	"github.com/mochisuna/linebot-sample/infrastructure/db"
)

func main() {
	// parse options
	path := flag.String("c", "_tools/local/config.toml", "config file")
	flag.Parse()

	// import config
	conf := &config.Config{}
	log.Println(*path)
	if err := config.New(conf, *path); err != nil {
		panic(err)
	}

	// init db connection
	// master db
	dbmClient, err := db.NewMySQL(&conf.DBMaster)
	if err != nil {
		panic(err)
	}
	defer dbmClient.Close()
	// slave db
	dbsClient, err := db.NewMySQL(&conf.DBSlave)
	if err != nil {
		panic(err)
	}
	defer dbsClient.Close()

	// initialize and injection relay
	// init repository
	eventRepo := infrastructure.NewEventRepository(dbmClient, dbsClient)
	ownerRepo := infrastructure.NewOwnerRepository(dbmClient, dbsClient)
	userRepo := infrastructure.NewUserRepository(dbmClient, dbsClient)
	// init application service
	callbackService := application.NewCallbackService(eventRepo, ownerRepo, userRepo)

	// inject all services
	services := &handler.Services{
		CallbackService: callbackService,
	}

	bot := handler.NewLineBot(&conf.Line)

	// Run Api server
	server := handler.New(conf.Server.Port, services, bot)
	log.Println("Start server")
	if err := server.ListenAndServe(); err != nil {
		panic(fmt.Sprintf("Failed ListenAndServe. err: %v", err))
	}
}
