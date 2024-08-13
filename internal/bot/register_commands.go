// This file holds all command registration
package bot

import (
	"github.com/bwmarrin/discordgo"
	"github.com/shininglegend/shieldbot/internal/commands"
)

const (
	// These are the names of the slash commands, to be consistent with the command handler.
	cmdPingType   = "pings"
	cmdConfigType = "config" // Subcommands elsewhere
	cmdIsolate    = "isolate"
	cmdRestore    = "restore"
)

func (b *Bot) registerCommands() error {
	b.registeredCommands = make(map[string]*discordgo.ApplicationCommand)

	commands := []*discordgo.ApplicationCommand{
		{
			Name:        cmdPingType,
			Description: "Responds with Pong!",
		},
		{
			Name:        cmdIsolate,
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
			Name:        cmdRestore,
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
		cmd, err := b.session.ApplicationCommandCreate(b.session.State.User.ID, "", v)
		if err != nil {
			return err
		}
		b.registeredCommands[cmd.Name] = cmd
	}
	// Now add the config one, so that the others are already registered
	v := &discordgo.ApplicationCommand{
		Name:        cmdConfigType,
		Description: "Configure bot settings",
		Options:     b.getConfigSubcommands(),
	}
	cmd, err := b.session.ApplicationCommandCreate(b.session.State.User.ID, "", v)
	if err != nil {
		return err
	}
	b.registeredCommands[cmd.Name] = cmd

	return nil
}

func (b *Bot) getConfigSubcommands() []*discordgo.ApplicationCommandOption {
	commandChoices := make([]*discordgo.ApplicationCommandOptionChoice, 0, len(b.registeredCommands))
	for name := range b.registeredCommands {
		commandChoices = append(commandChoices, &discordgo.ApplicationCommandOptionChoice{
			Name:  name,
			Value: name,
		})
	}

	return []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        commands.ViewPermName,
			Description: "View current command permissions. Admins override this.",
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        commands.AddPermName,
			Description: "Set permission for a command. Admins override this.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "command",
					Description: "The command to set permission for",
					Required:    true,
					Choices:     commandChoices,
				},
				{
					Type:        discordgo.ApplicationCommandOptionRole,
					Name:        "role",
					Description: "The role that can use the command",
					Required:    true,
				},
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        commands.RemovePermName,
			Description: "Remove permission for a command",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "command",
					Description: "The command to remove permission from",
					Required:    true,
					Choices:     commandChoices,
				},
				{
					Type:        discordgo.ApplicationCommandOptionRole,
					Name:        "role",
					Description: "The role to remove permission for",
					Required:    true,
				},
			},
		},
		{
			Name:        commands.SetIsolationRoleName,
			Description: "Set the isolation role for the guild",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionRole,
					Name:        "role",
					Description: "The role to use for isolation",
					Required:    true,
				},
			},
		},
	}
}
