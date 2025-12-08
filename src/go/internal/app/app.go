package app

import (
	"goservice/internal/config"
	"goservice/internal/db"
	"goservice/internal/rabbit"
	"goservice/internal/service"
)

func Run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	mysql, err := db.Connect(cfg)
	if err != nil {
		return err
	}
	cache := service.NewCache()
	repo := service.NewRepository(mysql)
	syncer := service.NewSyncer(cache, repo)
	go syncer.Start()

	consumer, err := rabbit.NewConsumer(cfg, service.NewHandler(cache, repo))
	if err != nil {
		return err
	}
	return consumer.Start()
}
