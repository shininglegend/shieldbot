package bot

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/shininglegend/shieldbot/pkg/utils"
)

const (
	// Options for the /log command
	actionVerbalWarn = "verbal_warn"
	actionBotWarn    = "bot_warn"
	actionTimeout    = "timeout"
	actionIsolate    = "isolate"
	actionKick       = "kick"
	actionBan        = "ban"
	actionOther      = "other"
)

func (b *Bot) handleLogging(s *discordgo.Session, i *discordgo.InteractionCreate) *discordgo.MessageEmbed {
	// Get the user and action from the interaction
	user := i.ApplicationCommandData().Options[0].UserValue(s)
	action := i.ApplicationCommandData().Options[1].StringValue()

	// Log the action
	errEmd := b.logAction(s, i, user, action)
	if errEmd != nil {
		return errEmd
	}

	// Return the response
	return utils.CreateEmbed("Logged action", fmt.Sprintf("Logged action for %v: %v", user.Mention(), action))
}

func (b *Bot) handleLoggingExternal(s *discordgo.Session, i *discordgo.InteractionCreate) *discordgo.MessageEmbed {
	// Get the user and action from the interaction
	user, err := s.User(i.ApplicationCommandData().Options[0].StringValue())
	if user == nil || err != nil {
		embed := utils.CreateEmbed("Error", err.Error())
		embed.Color = 0xFF0000 // Red
		embed.Timestamp = time.Now().Format(time.RFC3339)
		return embed
	}
	action := i.ApplicationCommandData().Options[1].StringValue()

	// Log the action
	errEmd := b.logAction(s, i, user, action)
	if errEmd != nil {
		return errEmd
	}

	// Return the response
	return utils.CreateEmbed("Logged action", fmt.Sprintf("Logged action for %v: %v", user.Mention(), action))
}

func (b *Bot) logAction(s *discordgo.Session, i *discordgo.InteractionCreate, user *discordgo.User, action string) *discordgo.MessageEmbed {
	// Get the guild
	guild, err := s.Guild(i.GuildID)
	if err != nil {
		return utils.CreateErrorEmbed(s, i, "Error getting guild", err)
	}

	// Get the mod log channel
	modLogChannelID, err := b.pm.GetLogChannelID(guild.ID)
	if err != nil {
		return utils.CreateErrorEmbed(s, i, "Error getting mod log channel", err)
	}

	// Set color based on action
	var color int
	switch action {
	case actionVerbalWarn:
		color = 0xFFFF00 // Yellow
	case actionBotWarn:
		color = 0xFFFF00 // Yellow
	case actionTimeout:
		color = 0xFFA500 // Orange
	case actionIsolate:
		color = 0xFFA500 // Orange
	case actionKick:
		color = 0xFF0000 // Red
	case actionBan:
		color = 0xFF0000 // Red
	case actionOther:
		color = 0x0000FF // Blue
	default:
		// Invalid action
		return utils.CreateErrorEmbed(s, i, "Invalid action", fmt.Errorf("invalid action: %v", action))
	}
	// Set the default reason
	reason := "*Reason not provided, and should be included below.*" // Default reason
	if len(i.ApplicationCommandData().Options) > 2 {
		reason = i.ApplicationCommandData().Options[2].StringValue()
	}

	// Create the embed
	embed := &discordgo.MessageEmbed{
		Title:       "Moderator Action Log",
		Description: fmt.Sprintf("Moderator %v took action on %v `%v`", i.Member.User.Mention(), user.Mention(), user.ID),
		Timestamp:   time.Now().Format(time.RFC3339),
		Color:       color,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Action",
				Value:  action,
				Inline: true,
			},
			{
				Name:   "Reason",
				Value:  reason,
				Inline: true,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Further details and file attachments may be added below",
		},
	}

	// Send the embed, with the id in the message to make it easier to find
	_, err = s.ChannelMessageSendComplex(modLogChannelID, &discordgo.MessageSend{
		Content: fmt.Sprintf("User ID: %v", user.ID),
		Embed:   embed,
	})
	if err != nil {
		return utils.CreateErrorEmbed(s, i, "Error sending mod log message", err)
	}
	return nil
}
