package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"

	"github.com/shininglegend/shieldbot/internal/bot"
	"github.com/shininglegend/shieldbot/internal/config"
	"github.com/shininglegend/shieldbot/internal/database"
	"github.com/shininglegend/shieldbot/pkg/utils"
)

func main() {
	var bot *bot.Bot
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic: %v", r)
			debug.PrintStack()
			time.Sleep(5 * time.Second) // Wait for 5 seconds before restarting

			// Notifying the developer
			if bot != nil && bot.Session != nil {
				utils.SendToDevChannelDMs(bot.Session, fmt.Sprintf("Recovered from panic: %v", r), 5)
			}
			main() // Restart the bot
		}
	}()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	db, err := database.New(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	defer db.Close()

	bot, err = bot.New(cfg.Token, db)
	if err != nil {
		log.Fatalf("Error creating bot: %v", err)
	}

	err = bot.Start()
	if err != nil {
		utils.SendToDevChannelDMs(bot.Session, fmt.Sprintf("Failed to start: %v", err.Error()), 5)
		log.Fatalf("Error starting bot: %v", err)
	}
	log.Printf("Bot is running. Press Ctrl-C to stop.")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	bot.Stop()
}
