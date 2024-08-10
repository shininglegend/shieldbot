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

// CreateEmbed creates a simple embed with a title and description
func CreateEmbed(title, description string) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title:       title,
		Description: description,
		Color:       0x00BFFF, // DeepSkyBlue
	}
}

// CreateErrorEmbed creates an embed with an error message
func CreateErrorEmbed(description string) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title:       "Error",
		Description: description,
		Color:       0xFF0000, // Red
	}
}
