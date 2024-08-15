// internal/bot/handlers.go
package bot

import (
	"fmt"
	"log"
	"strings"

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

	// Process the command
	var embed *discordgo.MessageEmbed
	switch n := i.ApplicationCommandData().Name; n {
	case cmdPingType:
		embed = b.handlePing(s, i)
	case cmdHelp:
		embed = b.handleHelp(s, i)
	case cmdConfigType: // All /config commands are delegated to the same function
		embed = b.pc.HandleConfig(s, i) // Needs admin permissions
	case cmdIsolate:
		embed = b.handleIsolate(s, i) // Needs manage roles permissions
	case cmdRestore:
		embed = b.handleRestore(s, i) // Needs manage roles permissions
	default:
		embed = utils.CreateNotAllowedEmbed("Unknown command", fmt.Sprintf("Unknown command: %v", n))
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

// Help command showing basic bot usage
func (b *Bot) handleHelp(_ *discordgo.Session, _ *discordgo.InteractionCreate) *discordgo.MessageEmbed {
	embed := utils.CreateEmbed("Bot Help", "This is a basic help command for the bot.")
	info := []*discordgo.MessageEmbedField{
		// Default commands
		{
			Name:   "Commands",
			Value:  "/ping - Check the bot's response time",
			Inline: false,
		},
		// Isolate and restore
		{
			Name: "Isolation Commands",
			Value: "/isolate - Isolate a user in the guild\n" +
				"/restore - Restore a user in the guild\n",
			Inline: false,
		},
		// Config commands
		{
			Name: "Config Commands",
			Value: "/config setisolationrole - Set the isolation role for the guild\n" +
				"/config viewperms - View the permissions of commands for the guild\n" +
				"/config addperm - Set the permission override for a command for a role\n" +
				"/config removeperm - Remove the permission override for a command for a role\n",
			Inline: false,
		},
	}
	embed.Fields = append(embed.Fields, info...)
	return embed
}

func (b *Bot) handleDMMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore messages from the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}

	// Check if the message is a DM and from the dev channel
	channel, err := s.Channel(m.ChannelID)
	if err != nil {
		log.Printf("Error getting channel: %v", err)
		return
	}

	if channel.Type != discordgo.ChannelTypeDM {
		return
	}

	// if the message isn't from the dev channel, send it to the dev channel
	if channel.ID != utils.DevChannelId {
		_, err = s.ChannelMessageSend(utils.DevChannelId, fmt.Sprintf("Message from %v (`%v`): %v", m.Author.Username, m.Author.ID, m.Content))
		if err != nil {
			log.Printf("Error sending message to dev channel: %v", err)
			err = nil
		}
		return
	}

	// Check if the message is "refresh"
	if strings.ToLower(m.Content) == "refresh" {
		err := b.RefreshCommands()
		if err != nil {
			log.Printf("Error refreshing commands: %v", err)
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Error refreshing commands: %v", err))
		} else {
			s.ChannelMessageSend(m.ChannelID, "Commands refreshed successfully!")
		}
	}
}
