// internal/commands/permcommands.go
package commands

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/shininglegend/shieldbot/internal/permissions"
	"github.com/shininglegend/shieldbot/pkg/utils"
)

const (
	ViewPermName         = "viewperms"
	AddPermName          = "addperm"
	RemovePermName       = "removeperm"
	SetIsolationRoleName = "setisolationrole"
)

type PermissionCommands struct {
	pm *permissions.PermissionManager
}

func NewPermissionCommands(pm *permissions.PermissionManager) *PermissionCommands {
	return &PermissionCommands{pm: pm}
}

func (pc *PermissionCommands) HandleConfig(s *discordgo.Session, i *discordgo.InteractionCreate) *discordgo.MessageEmbed {
	options := i.ApplicationCommandData().Options
	subcommand := options[0].Name

	switch subcommand {
	case ViewPermName:
		return pc.handleViewPerms(s, i)
	case AddPermName:
		return pc.handleAddPerm(s, i, options[0].Options)
	case RemovePermName:
		return pc.handleRemovePerm(s, i, options[0].Options)
	case SetIsolationRoleName:
		return pc.handleSetIsolationRole(s, i, options[0].Options)
	default:
		return utils.CreateErrorEmbed(fmt.Sprintf("Unknown subcommand: %v", subcommand))
	}
}

func (pc *PermissionCommands) handleViewPerms(s *discordgo.Session, in *discordgo.InteractionCreate) *discordgo.MessageEmbed {
	perms, err := pc.pm.GetCommandPermissions(in.GuildID)
	if err != nil {
		return utils.CreateErrorEmbed(fmt.Sprintf("Error retrieving permissions: %v", err))
	}

	var description strings.Builder
	if len(perms) == 0 {
		description.WriteString("No Guild permissions set.")
	}
	for command, roles := range perms {
		roleNames := make([]string, len(roles))
		for i, roleID := range roles {
			role, err := s.State.Role(in.GuildID, roleID)
			if err != nil {
				roleNames[i] = roleID // Use ID if role name can't be fetched
			} else {
				roleNames[i] = fmt.Sprintf("%v `%v`", role.Mention(), role.ID)
			}
		}
		description.WriteString(fmt.Sprintf("**%s**: %s\n", command, strings.Join(roleNames, ", ")))
	}

	return utils.CreateEmbed("Command Permissions", description.String())
}

func (pc *PermissionCommands) handleAddPerm(s *discordgo.Session, i *discordgo.InteractionCreate, options []*discordgo.ApplicationCommandInteractionDataOption) *discordgo.MessageEmbed {
	commandName := options[0].StringValue()
	role := options[1].RoleValue(s, i.GuildID)

	err := pc.pm.SetCommandPermission(i.GuildID, commandName, role.ID)
	if err != nil {
		return utils.CreateErrorEmbed(fmt.Sprintf("Error adding permission: %v", err))
	}
	return utils.CreateEmbed("Permission Added", fmt.Sprintf("Permission for command '%s' has been granted to role %s", commandName, role.Mention()))
}

func (pc *PermissionCommands) handleRemovePerm(s *discordgo.Session, i *discordgo.InteractionCreate, options []*discordgo.ApplicationCommandInteractionDataOption) *discordgo.MessageEmbed {
	if len(options) < 2 {
		return utils.CreateErrorEmbed("Not enough options provided. Please specify both command and role.")
	}

	commandName := options[0].StringValue()
	role := options[1].RoleValue(s, i.GuildID)

	if role == nil {
		return utils.CreateErrorEmbed("Invalid role provided.")
	}

	err := pc.pm.RemoveCommandPermission(i.GuildID, commandName, role.ID)
	if err != nil {
		return utils.CreateErrorEmbed(fmt.Sprintf("Error removing permission: %v", err))
	}

	return utils.CreateEmbed("Permission Removed", fmt.Sprintf("Permission for command '%s' has been removed from role %s", commandName, role.Mention()))
}

func (pc *PermissionCommands) handleSetIsolationRole(s *discordgo.Session, i *discordgo.InteractionCreate, options []*discordgo.ApplicationCommandInteractionDataOption) *discordgo.MessageEmbed {
	role := options[0].RoleValue(s, i.GuildID)

	err := pc.pm.SetIsolationRole(i.GuildID, role.ID)
	if err != nil {
		return utils.CreateErrorEmbed(fmt.Sprintf("Error setting isolation role: %v", err))
	}
	return utils.CreateEmbed("Isolation Role Set", fmt.Sprintf("Isolation role has been set to %s", role.Mention()))
}
