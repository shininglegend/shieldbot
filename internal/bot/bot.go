// internal/bot/bot.go
package bot

import (
	"database/sql"

	"github.com/bwmarrin/discordgo"
	"github.com/google/martian/v3/log"
	"github.com/shininglegend/shieldbot/internal/commands"
	"github.com/shininglegend/shieldbot/internal/permissions"
)

type Bot struct {
	session *discordgo.Session
	db      *sql.DB
	pm      *permissions.PermissionManager
	pc      *commands.PermissionCommands
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
	log.Infof("Bot is running as %s", b.session.State.User.Username)

	b.pc.RegisterCommands(b.session)

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
