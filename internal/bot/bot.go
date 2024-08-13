// internal/bot/bot.go
package bot

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/shininglegend/shieldbot/internal/commands"
	"github.com/shininglegend/shieldbot/internal/permissions"
)

const (
	dev_channel_id = "1272795752145354873" // Channel ID, not User ID
)

type Bot struct {
	session            *discordgo.Session
	db                 *sql.DB
	pm                 *permissions.PermissionManager
	pc                 *commands.PermissionCommands
	registeredCommands map[string]*discordgo.ApplicationCommand
}

func New(token string, db *sql.DB) (*Bot, error) {
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, err
	}

	pm := permissions.NewPermissionManager(db)
	err = pm.SetupTables()
	if err != nil {
		return nil, err
	}

	pc := commands.NewPermissionCommands(pm)

	bot := &Bot{
		session: session,
		db:      db,
		pm:      pm,
		pc:      pc,
	}

	session.AddHandler(bot.handleCommands)

	return bot, nil
}

func (b *Bot) Start() error {
	err := b.session.Open()
	if err != nil {
		return err
	}

	err = b.registerCommands()
	if err != nil {
		return err
	}
	b.Log("Bot has started")
	return err
}

func (b *Bot) Stop() {
	b.session.Close()
}

// This does it's best to deliver messages!
func (b *Bot) Log(msg string) {
	_, err := b.session.ChannelMessageSend(dev_channel_id, msg)
	if err != nil {
		log.Printf("Failed to send message to developer! Error: %v", err)
		// Retry a few times
		var errors []error
		errors = append(errors, err)
		err = nil
		for i := 0; i < 5; i++ {
			time.Sleep(5 * time.Second)
			_, err := b.session.ChannelMessageSend(dev_channel_id, msg)
			if err != nil {
				errors = append(errors, err)
				err = nil
				continue
			}
			// Try to send the errors as well
			time.Sleep(100 * time.Microsecond)
			_, _ = b.session.ChannelMessageSend(dev_channel_id, fmt.Sprintf("Previous messages failed. Errors %v", errors))
			return
		}
		log.Printf("Message: %v", msg)
		log.Fatalf("Errors: %v", errors)
	}
}
