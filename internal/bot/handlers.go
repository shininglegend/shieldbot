// internal/bot/handlers.go
package bot

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/shininglegend/shieldbot/pkg/utils"
)

func (b *Bot) handleCommands(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Acknowledge the interaction immediately
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	if err != nil {
		log.Printf("Error acknowledging interaction: %v", err)
		return
	}

	// Check if user is admin
	isAdmin, err := b.pm.IsAdmin(s, i.GuildID, i.Member.User.ID)
	if err != nil {
		log.Printf("Error checking admin status: %v", err)
		b.editResponseEmbed(s, i, utils.CreateErrorEmbed("An error occurred while checking permissions."))
		return
	}

	// If not admin, check regular permissions
	if !isAdmin {
		canUse, err := b.pm.CanUseCommand(i.GuildID, i.Member.User.ID, i.ApplicationCommandData().Name)
		if err != nil {
			log.Printf("Error checking permissions: %v", err)
			b.editResponseEmbed(s, i, utils.CreateErrorEmbed("An error occurred while checking permissions."))
			return
		}
		// Admin bypass
		if !canUse {
			b.editResponseEmbed(s, i, utils.CreateErrorEmbed("You don't have permission to use this command."))
			return
		}
	}

	// Process the command
	var embed *discordgo.MessageEmbed
	switch i.ApplicationCommandData().Name {
	case "ping":
		embed = b.handlePing(s, i)
	case "setperm":
		embed = b.pc.HandleSetPerm(s, i)
	case "setisolationrole":
		embed = b.pc.HandleSetIsolationRole(s, i)
	case "isolate":
		embed = b.handleIsolate(s, i)
	case "restore":
		embed = b.handleRestore(s, i)
	default:
		embed = utils.CreateErrorEmbed("Unknown command")
	}

	// Edit the original response with the command output
	b.editResponseEmbed(s, i, embed)
}

func (b *Bot) editResponseEmbed(s *discordgo.Session, i *discordgo.InteractionCreate, embed *discordgo.MessageEmbed) {
	_, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{embed},
	})
	if err != nil {
		log.Printf("Error editing interaction response: %v", err)
	}
}

func (b *Bot) handlePing(s *discordgo.Session, _ *discordgo.InteractionCreate) *discordgo.MessageEmbed {
	// Return the bot response time in milliseconds
	return utils.CreateEmbed("Pong!", fmt.Sprintf("%v ms", s.HeartbeatLatency().Milliseconds()))
}
