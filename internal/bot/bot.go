// internal/bot/bot.go
package bot

import (
	"database/sql"

	"github.com/bwmarrin/discordgo"
)

type Bot struct {
	session *discordgo.Session
	db      *sql.DB
}

func New(token string, db *sql.DB) (*Bot, error) {
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, err
	}

	bot := &Bot{
		session: session,
		db:      db,
	}

	session.AddHandler(bot.handleCommands)

	return bot, nil
}

func (b *Bot) Start() error {
	err := b.session.Open()
	if err != nil {
		return err
	}

	return b.registerCommands()
}

func (b *Bot) Stop() {
	b.session.Close()
}

func (b *Bot) registerCommands() error {
	commands := []*discordgo.ApplicationCommand{
		{
			Name:        "ping",
			Description: "Responds with Pong!",
		},
		{
			Name:        "isolate",
			Description: "Isolates a user by removing their roles",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionUser,
					Name:        "user",
					Description: "The user to isolate",
					Required:    true,
				},
			},
		},
		{
			Name:        "restore",
			Description: "Restores a user's roles",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionUser,
					Name:        "user",
					Description: "The user to restore",
					Required:    true,
				},
			},
		},
	}

	for _, v := range commands {
		_, err := b.session.ApplicationCommandCreate(b.session.State.User.ID, "", v)
		if err != nil {
			return err
		}
	}

	return nil
}
