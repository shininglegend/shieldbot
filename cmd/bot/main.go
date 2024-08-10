package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/shininglegend/shieldbot/internal/bot"
	"github.com/shininglegend/shieldbot/internal/config"
	"github.com/shininglegend/shieldbot/internal/database"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	db, err := database.New(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	defer db.Close()

	bot, err := bot.New(cfg.Token, db)
	if err != nil {
		log.Fatalf("Error creating bot: %v", err)
	}

	err = bot.Start()
	if err != nil {
		log.Fatalf("Error starting bot: %v", err)
	}

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	bot.Stop()
}
