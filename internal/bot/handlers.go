// internal/bot/handlers.go
package bot

import (
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func (b *Bot) handleCommands(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}

	switch i.ApplicationCommandData().Name {
	case "ping":
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Pong!",
			},
		})
	case "isolate":
		b.handleIsolate(s, i)
	case "restore":
		b.handleRestore(s, i)
	}
}

func (b *Bot) handleIsolate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options
	user := options[0].UserValue(s)

	member, err := s.GuildMember(i.GuildID, user.ID)
	if err != nil {
		log.Printf("Error fetching member: %v", err)
		respondWithError(s, i, "Failed to fetch member")
		return
	}

	// Save current roles
	roleIDs := strings.Join(member.Roles, ",")
	_, err = b.db.Exec("INSERT INTO user_roles (user_id, guild_id, roles) VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE roles = ?",
		user.ID, i.GuildID, roleIDs, roleIDs)
	if err != nil {
		log.Printf("Error saving roles: %v", err)
		respondWithError(s, i, "Failed to save roles")
		return
	}

	// Remove all roles
	for _, roleID := range member.Roles {
		err = s.GuildMemberRoleRemove(i.GuildID, user.ID, roleID)
		if err != nil {
			log.Printf("Error removing role: %v", err)
		}
	}

	// Add isolation role (assuming you have created this role and have its ID)
	isolationRoleID := "YOUR_ISOLATION_ROLE_ID"
	err = s.GuildMemberRoleAdd(i.GuildID, user.ID, isolationRoleID)
	if err != nil {
		log.Printf("Error adding isolation role: %v", err)
		respondWithError(s, i, "Failed to add isolation role")
		return
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("User %s has been isolated.", user.Username),
		},
	})
}

func (b *Bot) handleRestore(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options
	user := options[0].UserValue(s)

	var roleIDs string
	err := b.db.QueryRow("SELECT roles FROM user_roles WHERE user_id = ? AND guild_id = ?", user.ID, i.GuildID).Scan(&roleIDs)
	if err != nil {
		log.Printf("Error fetching roles: %v", err)
		respondWithError(s, i, "Failed to fetch roles")
		return
	}

	// Remove isolation role
	isolationRoleID := "YOUR_ISOLATION_ROLE_ID"
	err = s.GuildMemberRoleRemove(i.GuildID, user.ID, isolationRoleID)
	if err != nil {
		log.Printf("Error removing isolation role: %v", err)
	}

	// Restore old roles
	for _, roleID := range strings.Split(roleIDs, ",") {
		err = s.GuildMemberRoleAdd(i.GuildID, user.ID, roleID)
		if err != nil {
			log.Printf("Error adding role: %v", err)
		}
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("Roles for user %s have been restored.", user.Username),
		},
	})
}

func respondWithError(s *discordgo.Session, i *discordgo.InteractionCreate, message string) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: message,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}
