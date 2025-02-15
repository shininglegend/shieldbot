package utils

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type Messages []string

func (m Messages) GetMessages(original string) string {
	// Combine the messages with a newline character, or return an empty string if there are no messages
	if len(m) == 0 {
		return original
	}
	var i string
	for _, message := range m {
		i = fmt.Sprintf("%v\n%v", i, message)
	}
	return fmt.Sprintf("%v\n%v", original, i)
}

func (m *Messages) AddMessage(message string) {
	*m = append(*m, message)
}

// SafeRoleName returns a string representation of the role that won't ping members
func SafeRoleName(role *discordgo.Role) string {
	return strings.ReplaceAll(role.Name, "@", "@\u200B")
}

// SafeUser returns a User consistently
func SafeUser(interaction *discordgo.Interaction) *discordgo.User {
	if interaction.Member != nil {
		return interaction.Member.User
	}
	return interaction.User
}

// Checks an error message for a specific discord error code
func CheckError(err error, checkCode int) bool {
	e, ok := err.(*discordgo.RESTError)
	return ok && e.Message != nil && e.Message.Code == checkCode
}

// CreateEmbed creates a simple embed with a title and description
func CreateEmbed(title, description string) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title:       title,
		Description: description,
		Color:       0x00FF00, // Green
	}
}

// CreateErrorEmbed i *discordgo.InteractionCreate, creates an embed with an error message, and dms the developer
func CreateErrorEmbed(s *discordgo.Session, i *discordgo.InteractionCreate, desc string, err error) *discordgo.MessageEmbed {
	// Send a message to the developer with a link to the message, the user, and the error
	SendToDevChannelDMs(s, fmt.Sprintf("Error: `%v`\nUser: %v | Channel: %v", err, SafeUser(i.Interaction).ID, i.ChannelID), 2)
	return &discordgo.MessageEmbed{
		Title:       "Error",
		Description: desc,
		Color:       0xFF0000, // Red
		Footer: &discordgo.MessageEmbedFooter{
			Text:         "Error reported to developer",
			IconURL:      "",
			ProxyIconURL: "",
		},
	}
}

// CreateNotAllowedEmbed creates an embed with a permission error message in yellow
func CreateNotAllowedEmbed(title, desc string) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title:       title,
		Description: desc,
		Color:       0xFFFF00, // Yellow
	}
}

// Contails function
func Contains(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

func Remove(slice []string, item string) []string {
	for i, v := range slice {
		if v == item {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}

// Helper function to get the highest role, could return nil
func GetHighestRole(memberRoles []string, guildRoles []*discordgo.Role) *discordgo.Role {
	var highestRole *discordgo.Role
	for _, roleID := range memberRoles {
		for _, guildRole := range guildRoles {
			if guildRole.ID == roleID {
				if highestRole == nil || guildRole.Position > highestRole.Position {
					highestRole = guildRole
				}
				//break
			}
		}
	}
	return highestRole
}
