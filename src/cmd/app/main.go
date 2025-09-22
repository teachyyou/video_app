package main

import (
	"awesomeProject/src/app/config"
	"awesomeProject/src/app/handler"
	"awesomeProject/src/app/repository"
	"awesomeProject/src/app/server"
	"awesomeProject/src/app/service"

	"github.com/joho/godotenv"
	"go.uber.org/fx"
)

func init() {
	_ = godotenv.Load()
}

func main() {

	fx.New(
		server.Module,
		handler.HelloModule,
		handler.VideoModule,
		config.Module,
		config.DbModule,
		service.VideoModule,
		repository.VideoRepoModule,
	).Run()
}
