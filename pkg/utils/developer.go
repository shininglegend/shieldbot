// This file handles Developer helpers for better debugging and development.
package utils

import (
	"fmt"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
)

const (
	DeveloperID = "585991293377839114" // User ID
)

var DevChannel = (*discordgo.Channel)(nil)

// This does it's best to deliver messages!
func SendToDevChannelDMs(s *discordgo.Session, msg string, retries int) {
	_, err := s.ChannelMessageSend(GetDevChannel(s), msg)
	if err != nil {
		log.Printf("Failed to send message to developer! Error: %v", err)
		// Retry a few times
		var errors []error
		errors = append(errors, err)
		err = nil
		for i := 1; i < retries; i++ {
			time.Sleep(5 * time.Second)
			_, err := s.ChannelMessageSend(GetDevChannel(s), msg)
			if err != nil {
				errors = append(errors, err)
				err = nil
				continue
			}
			// Try to send the errors as well
			time.Sleep(100 * time.Microsecond)
			_, _ = s.ChannelMessageSend(GetDevChannel(s), fmt.Sprintf("Previous messages failed. Errors %v", errors))
			return
		}
		log.Printf("Message: %v", msg)
		log.Fatalf("Errors: %v", errors)
	}
}

// Get the channels for DMs with the developer based on the developer's hard-coded ID
func GetDevChannel(s *discordgo.Session) string {
	if DevChannel != nil {
		return DevChannel.ID
	}
	dev, err := s.User(DeveloperID)
	if err != nil {
		log.Fatalf("Error getting developer: %v", err)
	}
	ch, err := s.UserChannelCreate(dev.ID)
	if err != nil {
		log.Fatalf("Error creating DM channel with developer: %v", err)
	}
	DevChannel = ch
	return ch.ID
}
