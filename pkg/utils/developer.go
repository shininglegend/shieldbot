// This file handles Developer helpers for better debugging and development.
package utils

import (
	"fmt"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
)

const (
	DevChannelId = "1272795752145354873" // Channel ID, not User ID
)

// This does it's best to deliver messages!
func SendToDevChannelDMs(s *discordgo.Session, msg string, retries int) {
	_, err := s.ChannelMessageSend(DevChannelId, msg)
	if err != nil {
		log.Printf("Failed to send message to developer! Error: %v", err)
		// Retry a few times
		var errors []error
		errors = append(errors, err)
		err = nil
		for i := 1; i < retries; i++ {
			time.Sleep(5 * time.Second)
			_, err := s.ChannelMessageSend(DevChannelId, msg)
			if err != nil {
				errors = append(errors, err)
				err = nil
				continue
			}
			// Try to send the errors as well
			time.Sleep(100 * time.Microsecond)
			_, _ = s.ChannelMessageSend(DevChannelId, fmt.Sprintf("Previous messages failed. Errors %v", errors))
			return
		}
		log.Printf("Message: %v", msg)
		log.Fatalf("Errors: %v", errors)
	}
}
