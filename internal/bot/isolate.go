package bot

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/shininglegend/shieldbot/pkg/auth"
	"github.com/shininglegend/shieldbot/pkg/utils"
)

func (b *Bot) handleIsolate(s *discordgo.Session, i *discordgo.InteractionCreate) *discordgo.MessageEmbed {
	options := i.ApplicationCommandData().Options
	user := options[0].UserValue(s)
	messages := utils.Messages{}

	// Authorize the command
	if e := auth.QuickAuthManageRolesOrOverride(b.pm, s, i); e != nil {
		return e
	}

	// Ensure the person issuing the command has a role that is higher than the target's highest
	issuer, err := s.GuildMember(i.GuildID, i.Member.User.ID)
	if err != nil {
		log.Printf("Error fetching issuer member: %v", err)
		return utils.CreateErrorEmbed(s, "Failed to fetch issuer member", err)
	}

	guild, err := s.Guild(i.GuildID)
	if err != nil {
		log.Printf("Error fetching guild: %v", err)
		return utils.CreateErrorEmbed(s, "Failed to fetch guild", err)
	}

	member, err := s.GuildMember(i.GuildID, user.ID)
	if err != nil {
		log.Printf("Error fetching member: %v", err)
		return utils.CreateErrorEmbed(s, "Failed to fetch member", err)
	}
	// Get isolation role
	isolationRoleID, err := b.pm.GetIsolationRoleID(i.GuildID)
	if err != nil {
		if err == sql.ErrNoRows {
			return utils.CreateNotAllowedEmbed("Isolation role not set.", "Please set it using /config setisolationrole.")
		}
		log.Printf("Error fetching isolation role: %v", err)
		return utils.CreateErrorEmbed(s, "Failed to fetch isolation role", err)
	}

	issuerHighestRole := getHighestRole(issuer.Roles, guild.Roles)
	targetHighestRole := getHighestRole(member.Roles, guild.Roles)

	if issuerHighestRole.Position <= targetHighestRole.Position {
		return utils.CreateNotAllowedEmbed("Ayo, you can't do that!", "You don't have permission to isolate this user. Your highest role must be higher than the target user's highest role.")
	}

	// Save current roles
	var roleIDs string
	for _, roleID := range member.Roles {
		if roleID == isolationRoleID {
			return utils.CreateNotAllowedEmbed("Already done!", fmt.Sprintf("User %s is already isolated.", user.Mention()))
		}
		if roleIDs != "" {
			roleIDs = fmt.Sprintf("%s,%s", roleIDs, roleID)
		} else {
			roleIDs = roleID
		}
	}
	_, err = b.db.Exec(`
        INSERT INTO user_roles (user_id, guild_id, roles) 
        VALUES (?, ?, ?) 
        ON CONFLICT(user_id, guild_id) 
        DO UPDATE SET roles = ?`,
		user.ID, i.GuildID, roleIDs, roleIDs)
	if err != nil {
		log.Printf("Error saving roles: %v", err)
		return utils.CreateErrorEmbed(s, "Error: Failed to save roles to database. Manually isolate the user.", err)
	}

	// Remove all roles
	for _, roleID := range member.Roles {
		if roleID == isolationRoleID {
			continue
		}
		err = s.GuildMemberRoleRemove(i.GuildID, user.ID, roleID)
		if err != nil {
			log.Printf("Error removing role: %v", err)
			messages.AddMessage(fmt.Sprintf("Failed to remove role %s: %v", roleID, err.Error()))
			continue
		}
		// Fetch the role
		rolemsg := fmt.Sprintf("`%v`", roleID)
		role, err := s.State.Role(i.GuildID, roleID)
		if err != nil {
			log.Printf("Error fetching role: %v", err)
			err = nil
		} else {
			rolemsg = role.Mention()
		}
		messages.AddMessage(fmt.Sprintf("Removed role %v from %v", rolemsg, user.Mention()))
	}

	// Add isolation role based on db
	err = s.GuildMemberRoleAdd(i.GuildID, user.ID, isolationRoleID)
	if err != nil {
		log.Printf("Error adding isolation role: %v", err)
		return utils.CreateErrorEmbed(s, "Failed to add isolation role", err)
	}
	return utils.CreateEmbed(fmt.Sprintf("User %s (`%v`) has been isolated.", user.Username, user.ID), messages.GetMessages(""))
}

func (b *Bot) handleRestore(s *discordgo.Session, i *discordgo.InteractionCreate) *discordgo.MessageEmbed {
	options := i.ApplicationCommandData().Options
	user := options[0].UserValue(s)
	messages := utils.Messages{}

	// Authorize the command
	if e := auth.QuickAuthManageRolesOrOverride(b.pm, s, i); e != nil {
		return e
	}

	var roleIDs string
	err := b.db.QueryRow("SELECT roles FROM user_roles WHERE user_id = ? AND guild_id = ?", user.ID, i.GuildID).Scan(&roleIDs)
	if err != nil {
		if err == sql.ErrNoRows {
			return utils.CreateNotAllowedEmbed("Unable to restore", "No roles found to restore. Are you sure this user was isolated using the bot?")
		} else {
			log.Printf("Error fetching roles: %v", err)
			return utils.CreateErrorEmbed(s, "Failed to fetch roles", err)
		}
	}

	// Get isolation role
	isolationRoleID, err := b.pm.GetIsolationRoleID(i.GuildID)
	if err != nil {
		if err == sql.ErrNoRows {
			return utils.CreateNotAllowedEmbed("Isolation role not set.", "Please set it using /setisolationrole.")
		}
		log.Printf("Error fetching isolation role: %v", err)
		return utils.CreateErrorEmbed(s, "Failed to fetch isolation role", err)
	}

	// Ensure the person issuing the command has a role that is higher than the target's highest
	issuer, err := s.GuildMember(i.GuildID, i.Member.User.ID)
	if err != nil {
		log.Printf("Error fetching issuer member: %v", err)
		return utils.CreateErrorEmbed(s, "Failed to fetch issuer member", err)
	}

	guild, err := s.Guild(i.GuildID)
	if err != nil {
		log.Printf("Error fetching guild: %v", err)
		return utils.CreateErrorEmbed(s, "Failed to fetch guild", err)
	}

	member, err := s.GuildMember(i.GuildID, user.ID)
	if err != nil {
		log.Printf("Error fetching member: %v", err)
		return utils.CreateErrorEmbed(s, "Failed to fetch member", err)
	}

	issuerHighestRole := getHighestRole(issuer.Roles, guild.Roles)
	targetHighestRole := getHighestRole(member.Roles, guild.Roles)

	if issuerHighestRole.Position <= targetHighestRole.Position {
		return utils.CreateErrorEmbed(s, "You don't have permission to isolate this user. Your highest role must be higher than the target user's highest role.", err)
	}

	// Check if user is isolated
	for _, roleID := range member.Roles {
		if roleID == isolationRoleID {
			goto pass
		}
	}
	return utils.CreateErrorEmbed(s, fmt.Sprintf("User %s is not isolated.", user.Mention()), err)

pass:
	// Remove isolation role
	err = s.GuildMemberRoleRemove(i.GuildID, user.ID, isolationRoleID)
	if err != nil {
		log.Printf("Error removing isolation role: %v", err)
		messages.AddMessage("Failed to remove isolation role")
	}

	// Restore old roles
	if roleIDs != "" {
		for _, roleID := range strings.Split(roleIDs, ",") {
			if roleID == "" {
				continue
			}
			err = s.GuildMemberRoleAdd(i.GuildID, user.ID, roleID)
			if err != nil {
				log.Printf("Error adding role: %v", err)
				messages.AddMessage(fmt.Sprintf("Failed to add role %s: %v", roleID, err.Error()))
				continue
			}
			// Fetch the roles
			rolemsg := fmt.Sprintf("`%v`", roleID)
			role, err := s.State.Role(i.GuildID, roleID)
			if err != nil {
				log.Printf("Error fetching role: %v", err)
				err = nil
			} else {
				rolemsg = role.Mention()
			}
			messages.AddMessage(fmt.Sprintf("Restored role %v to %v", rolemsg, user.Mention()))
		}
		// Log the restored roles
		log.Printf("Restored roles for user %s: %s", user.Username, roleIDs)
	}

	// Delete the user's roles from the database
	_, err = b.db.Exec("DELETE FROM user_roles WHERE user_id = ? AND guild_id = ?", user.ID, i.GuildID)
	if err != nil {
		log.Printf("Error deleting roles: %v", err)
	}
	return utils.CreateEmbed(fmt.Sprintf("User %s (`%v`) has been restored.", user.Username, user.ID), messages.GetMessages(""))
}

// Helper function to get the highest role
func getHighestRole(memberRoles []string, guildRoles []*discordgo.Role) *discordgo.Role {
	var highestRole *discordgo.Role
	for _, roleID := range memberRoles {
		for _, guildRole := range guildRoles {
			if guildRole.ID == roleID {
				if highestRole == nil || guildRole.Position > highestRole.Position {
					highestRole = guildRole
				}
				break
			}
		}
	}
	return highestRole
}
