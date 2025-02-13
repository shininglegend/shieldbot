// internal/commands/permcommands.go
package commands

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/shininglegend/shieldbot/internal/permissions"
	"github.com/shininglegend/shieldbot/pkg/auth"
	"github.com/shininglegend/shieldbot/pkg/utils"
)

const (
	ViewPermName         = "viewperms"
	AddPermName          = "addperm"
	RemovePermName       = "removeperm"
	SetIsolationRoleName = "setisolationrole"
	SetLogChannel        = "setlogchannel"
)

type PermissionCommands struct {
	pm *permissions.PermissionManager
}

func NewPermissionCommands(pm *permissions.PermissionManager) *PermissionCommands {
	return &PermissionCommands{pm: pm}
}

func (pc *PermissionCommands) HandleConfig(s *discordgo.Session, i *discordgo.InteractionCreate) *discordgo.MessageEmbed {
	// Perms check
	if embed := auth.QuickAuthAdminOrOverride(pc.pm, s, i); embed != nil {
		return embed
	}
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
	case SetLogChannel:
		return pc.handleSetLogChannel(s, i, options[0].Options)
	default:
		return utils.CreateNotAllowedEmbed("Unknown subcommand to config", fmt.Sprintf("Unknown subcommand: %v", subcommand))
	}
}

func (pc *PermissionCommands) handleViewPerms(s *discordgo.Session, in *discordgo.InteractionCreate) *discordgo.MessageEmbed {
	perms, err := pc.pm.GetCommandPermissions(in.GuildID)
	if err != nil {
		return utils.CreateErrorEmbed(s, in, "Error retrieving permissions", err)
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

	return utils.CreateEmbed("Command Permission Overrides:", description.String())
}

func (pc *PermissionCommands) handleAddPerm(s *discordgo.Session, i *discordgo.InteractionCreate, options []*discordgo.ApplicationCommandInteractionDataOption) *discordgo.MessageEmbed {
	commandName := options[0].StringValue()
	role := options[1].RoleValue(s, i.GuildID)

	err := pc.pm.SetCommandPermission(i.GuildID, commandName, role.ID)
	if err != nil {
		return utils.CreateErrorEmbed(s, i, "Error adding permission", err)
	}
	return utils.CreateEmbed("Permission Added", fmt.Sprintf("Permission for command '%s' has been granted to role %s", commandName, role.Mention()))
}

func (pc *PermissionCommands) handleRemovePerm(s *discordgo.Session, i *discordgo.InteractionCreate, options []*discordgo.ApplicationCommandInteractionDataOption) *discordgo.MessageEmbed {
	if len(options) < 2 {
		return utils.CreateNotAllowedEmbed("Not enough options provided.", "Please specify both command and role.")
	}
	commandName := options[0].StringValue()
	role := options[1].RoleValue(s, i.GuildID)
	// Check for existance
	perms, err := pc.pm.GetCommandPermissions(i.GuildID)
	if err != nil {
		return utils.CreateErrorEmbed(s, i, "Error retrieving permissions", err)
	}
	// Check if the role has overrides
	if _, ok := perms[commandName]; !ok {
		return utils.CreateNotAllowedEmbed("No permission override found", "No permission override found for the command.")
	}
	// Check if the role has the permission
	if !utils.Contains(perms[commandName], role.ID) {
		return utils.CreateNotAllowedEmbed("Role does not have permission", "Role does not have permission for the command.")
	}

	if role == nil {
		return utils.CreateNotAllowedEmbed("Role not found", "Role not found in the server.")
	}

	err = pc.pm.RemoveCommandPermission(i.GuildID, commandName, role.ID)
	if err != nil {
		return utils.CreateErrorEmbed(s, i, "Error removing permission", err)
	}

	return utils.CreateEmbed("Permission Removed", fmt.Sprintf("Permission for command '%s' has been removed from role %s", commandName, role.Mention()))
}

func (pc *PermissionCommands) handleSetIsolationRole(s *discordgo.Session, i *discordgo.InteractionCreate, options []*discordgo.ApplicationCommandInteractionDataOption) *discordgo.MessageEmbed {
	role := options[0].RoleValue(s, i.GuildID)
	// Ensure the role exists and you have perms to manage it
	if role == nil {
		return utils.CreateNotAllowedEmbed("Error setting isolation role", "The specified role does not exist")
	}

	// Get the bot's permissions in the guild
	botPerms, err := s.State.UserChannelPermissions(s.State.User.ID, i.ChannelID)
	if err != nil {
		return utils.CreateErrorEmbed(s, i, "Error checking bot permissions", err)
	}

	// Check if the bot has permission to manage roles
	if botPerms&discordgo.PermissionManageRoles == 0 {
		return utils.CreateNotAllowedEmbed("Insufficient bot permissions", "The bot doesn't have permission to manage roles")
	}

	// Get the guild and bot member information
	guild, err := s.State.Guild(i.GuildID)
	if err != nil {
		return utils.CreateErrorEmbed(s, i, "Error fetching guild information", err)
	}

	botMember, err := s.State.Member(i.GuildID, s.State.User.ID)
	if err != nil {
		return utils.CreateErrorEmbed(s, i, "Error fetching bot member information", err)
	}

	// Use the helper function to get the bot's highest role
	highestBotRole := utils.GetHighestRole(botMember.Roles, guild.Roles)

	// Check if the bot's highest role is above the isolation role
	if highestBotRole == nil || highestBotRole.Position <= role.Position {
		return utils.CreateNotAllowedEmbed("Insufficient bot role hierarchy", "The bot's highest role is not above the specified isolation role")
	}

	err = pc.pm.SetIsolationRole(i.GuildID, role.ID)
	if err != nil {
		return utils.CreateErrorEmbed(s, i, "Error setting isolation role", err)
	}
	return utils.CreateEmbed("Isolation Role Set", fmt.Sprintf("Isolation role has been set to %s", role.Mention()))
}

func (pc *PermissionCommands) handleSetLogChannel(s *discordgo.Session, i *discordgo.InteractionCreate, options []*discordgo.ApplicationCommandInteractionDataOption) *discordgo.MessageEmbed {
	channel := options[0].ChannelValue(s)
	// Ensure the channel exists and you have perms to manage it
	if channel == nil {
		return utils.CreateNotAllowedEmbed("Error setting log channel", "The specified channel does not exist")
	}

	// Get the bot's permissions in the guild
	botPerms, err := s.State.UserChannelPermissions(s.State.User.ID, i.ChannelID)
	if err != nil {
		return utils.CreateErrorEmbed(s, i, "Error checking bot permissions", err)
	}

	// Check if the bot has permission to send messages in the channel
	if botPerms&discordgo.PermissionSendMessages == 0 {
		return utils.CreateNotAllowedEmbed("Insufficient bot permissions", "The bot doesn't have permission to send messages in the channel")
	}

	err = pc.pm.SetLogChannel(i.GuildID, channel.ID)
	if err != nil {
		return utils.CreateErrorEmbed(s, i, "Error setting log channel", err)
	}
	return utils.CreateEmbed("Log Channel Set", fmt.Sprintf("Log channel has been set to %s", channel.Mention()))
}
