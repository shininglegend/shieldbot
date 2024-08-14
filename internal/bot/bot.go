// internal/bot/bot.go
package bot

import (
	"database/sql"

	"github.com/bwmarrin/discordgo"
	"github.com/shininglegend/shieldbot/internal/commands"
	"github.com/shininglegend/shieldbot/internal/permissions"
	"github.com/shininglegend/shieldbot/pkg/utils"
)

type Bot struct {
	Session            *discordgo.Session
	db                 *sql.DB
	pm                 *permissions.PermissionManager
	pc                 *commands.PermissionCommands
	registeredCommands map[string]*discordgo.ApplicationCommand
}

func (*Bot) New(token string, db *sql.DB) (*Bot, error) {
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
		Session: session,
		db:      db,
		pm:      pm,
		pc:      pc,
	}

	session.AddHandler(bot.handleCommands)

	return bot, nil
}

func (b *Bot) Start() error {
	err := b.Session.Open()
	if err != nil {
		return err
	}

	err = b.registerCommands()
	if err != nil {
		return err
	}

	b.AddMessageHandlers()

	
	utils.SendToDevChannelDMs(b.Session, "Bot has started", 0)
	return err
}

func (b *Bot) Stop() {
	b.Session.Close()
}

func (b *Bot) AddMessageHandlers() {
	b.Session.AddHandler(b.handleDMMessage)
}
