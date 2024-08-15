package auth

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/shininglegend/shieldbot/internal/permissions"
	"github.com/shininglegend/shieldbot/pkg/utils"
)

const (
	ErrServerOnly = "This command is to be used in servers only."
	ErrorGeneric  = "An error occurred while processing your request, the develpers have been notified."
	ErrorNoPerms  = "You don't have permission to use this command."
)

// Check if a user has the manage roles permission or has an override for this command.
// Uses cache if possible.
func QuickAuthManageRolesOrOverride(pm *permissions.PermissionManager, s *discordgo.Session, i *discordgo.InteractionCreate) *discordgo.MessageEmbed {
	// Sanity check
	if i.Member == nil {
		return utils.CreateNotAllowedEmbed(ErrServerOnly, "What, you think I live here too?")
	}

	// Ensure the member has manage roles permission
	permissions, err := s.UserChannelPermissions(i.Member.User.ID, i.ChannelID)
	if err != nil {
		log.Printf("Error fetching permissions: %v", err)
		return utils.CreateErrorEmbed(s, i, ErrorGeneric, err)
	}
	if permissions&discordgo.PermissionManageRoles != discordgo.PermissionManageRoles {
		allowed, err := pm.CanUseCommand(s, i.GuildID, i.Member.User.ID, "isolate")
		if err != nil {
			log.Printf("Error checking permissions: %v", err)
			return utils.CreateErrorEmbed(s, i, ErrorGeneric, err)
		}
		if !allowed {
			return AddMissingPerms(utils.CreateNotAllowedEmbed(ErrorNoPerms, ""), []string{"Manage Roles"})
		}
	}
	return nil
}

// Check if a user has the admin permission or has an override for this command.
// Uses cache if possible.
func QuickAuthAdminOrOverride(pm *permissions.PermissionManager, s *discordgo.Session, i *discordgo.InteractionCreate) *discordgo.MessageEmbed {
	// Sanity check
	if i.Member == nil {
		return utils.CreateNotAllowedEmbed(ErrServerOnly, "What, you think I live here too?")
	}
	// Check if user is admin
	isAdmin, err := pm.IsAdmin(s, i.GuildID, i.Member.User.ID)
	if err != nil {
		return utils.CreateErrorEmbed(s, i, fmt.Sprintf("Error checking admin status: %v", err), err)
	}
	if !isAdmin {
		canUse, err := pm.CanUseCommand(s, i.GuildID, i.Member.User.ID, i.ApplicationCommandData().Name)
		if err != nil {
			log.Printf("Error checking permissions: %v", err)
			return utils.CreateErrorEmbed(s, i, ErrorGeneric, err)
		}
		if !canUse {
			return AddMissingPerms(utils.CreateNotAllowedEmbed(ErrorNoPerms, ""), []string{"Administrator"})
		}
	}
	return nil
}

// Check if a user has the manage messages permission or has an override for this command.
// Uses cache if possible.
func QuickAuthManageMessagesOrOverride(pm *permissions.PermissionManager, s *discordgo.Session, i *discordgo.InteractionCreate) *discordgo.MessageEmbed {
	// Sanity check
	if i.Member == nil {
		return utils.CreateNotAllowedEmbed(ErrServerOnly, "What, you think I live here too?")
	}
	// Ensure the member has manage messages permission
	permissions, err := s.UserChannelPermissions(i.Member.User.ID, i.ChannelID)
	if err != nil {
		log.Printf("Error fetching permissions: %v", err)
		return utils.CreateErrorEmbed(s, i, "Failed to fetch permissions", err)
	}
	if permissions&discordgo.PermissionManageMessages != discordgo.PermissionManageMessages {
		allowed, err := pm.CanUseCommand(s, i.GuildID, i.Member.User.ID, "purge")
		if err != nil {
			log.Printf("Error checking permissions: %v", err)
			return utils.CreateErrorEmbed(s, i, ErrorGeneric, err)
		}
		if !allowed {
			return AddMissingPerms(utils.CreateNotAllowedEmbed(ErrorNoPerms, ""), []string{"Manage Messages"})
		}
	}
	return nil
}

// Add the relevant missing perms to a message
func AddMissingPerms(embed *discordgo.MessageEmbed, missingPerms []string) *discordgo.MessageEmbed {
	if len(missingPerms) == 0 {
		return embed
	}
	var perms string
	for _, perm := range missingPerms {
		perms = fmt.Sprintf("%v\n%v", perms, perm)
	}
	embed.Description = fmt.Sprintf("%v\nMissing permissions:%v", embed.Description, perms)
	return embed
}
