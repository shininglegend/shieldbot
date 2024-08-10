// internal/commands/permcommands.go
package commands

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/shininglegend/shieldbot/internal/permissions"
	"github.com/shininglegend/shieldbot/pkg/utils"
)

type PermissionCommands struct {
	pm *permissions.PermissionManager
}

func NewPermissionCommands(pm *permissions.PermissionManager) *PermissionCommands {
	return &PermissionCommands{pm: pm}
}

func (pc *PermissionCommands) RegisterCommands(s *discordgo.Session) {
	commands := []*discordgo.ApplicationCommand{
		{
			Name:        "setperm",
			Description: "Set permission for a command",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "command",
					Description: "The command to set permission for",
					Required:    true,
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
			Name:        "setisolationrole",
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

	for _, v := range commands {
		_, err := s.ApplicationCommandCreate(s.State.User.ID, "", v)
		if err != nil {
			fmt.Printf("Cannot create '%v' command: %v", v.Name, err)
		}
	}
}

func (pc *PermissionCommands) HandleSetPerm(s *discordgo.Session, i *discordgo.InteractionCreate) *discordgo.MessageEmbed {
	options := i.ApplicationCommandData().Options
	commandName := options[0].StringValue()
	role := options[1].RoleValue(s, i.GuildID)

	err := pc.pm.SetCommandPermission(i.GuildID, commandName, role.ID)
	if err != nil {
		return utils.CreateErrorEmbed(fmt.Sprintf("Error setting permission: %v", err))
	}
	return utils.CreateEmbed("Permission Set", fmt.Sprintf("Permission for command '%s' has been granted to role %s", commandName, utils.SafeRoleName(role)))
}

func (pc *PermissionCommands) HandleSetIsolationRole(s *discordgo.Session, i *discordgo.InteractionCreate) *discordgo.MessageEmbed {
	options := i.ApplicationCommandData().Options
	role := options[0].RoleValue(s, i.GuildID)

	err := pc.pm.SetIsolationRole(i.GuildID, role.ID)
	if err != nil {
		return utils.CreateErrorEmbed(fmt.Sprintf("Error setting isolation role: %v", err))
	}
	return utils.CreateEmbed("Isolation Role Set", fmt.Sprintf("Isolation role has been set to %s", utils.SafeRoleName(role)))
}
