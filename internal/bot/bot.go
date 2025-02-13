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

	// Request intents
	session.Identify.Intents = discordgo.IntentsAllWithoutPrivileged | discordgo.IntentsGuildMembers

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
	// Open the session
	err := b.Session.Open()
	if err != nil {
		return err
	}
	// Register all other events
	err = b.registerCommands()
	if err != nil {
		return err
	}

	// Register join and leave events
	b.registerEvents()

	// Add message handlers
	b.AddMessageHandlers()

	utils.SendToDevChannelDMs(b.Session, "Bot has started", 0)
	return nil
}

func (b *Bot) Stop() {
	b.Session.Close()
}

func (b *Bot) AddMessageHandlers() {
	b.Session.AddHandler(b.handleDMMessage)
}

func (b *Bot) registerEvents() {
	b.Session.AddHandler(b.HandleJoin)
	b.Session.AddHandler(b.HandleLeave)
}
