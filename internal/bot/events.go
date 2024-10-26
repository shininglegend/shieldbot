package bot

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/shininglegend/shieldbot/pkg/utils"
)

// HandleJoin processes new member join events
func (b *Bot) HandleJoin(s *discordgo.Session, m *discordgo.GuildMemberAdd) {
	// If data is missing or they're a bot, ignore it.
	if m.User.Bot || m.Member == nil || m.Member.User == nil || m.Member.User.ID == "" {
		return
	}
	// Check if the user was isolated
	var roleIDs string
	err := b.db.QueryRow("SELECT roles FROM user_roles WHERE user_id = ? AND guild_id = ?", m.Member.User.ID, m.GuildID).Scan(&roleIDs)
	if err != nil {
		if err == sql.ErrNoRows {
			// If they weren't isolated, do nothing
			return
		} else {
			log.Printf("Error fetching roles: %v", err)
			utils.SendToDevChannelDMs(s, fmt.Sprintf("Error fetching roles: %v", err), 1)
		}
	}

	// If they were isolated, remove their roles and add the isolation role back
	isolationRole, err := b.pm.GetIsolationRoleID(m.GuildID)
	if err != nil {
		// PM the dev
		utils.SendToDevChannelDMs(s, fmt.Sprintf("Error fetching isolation role: %v", err), 1)
	}
	for _, role := range m.Member.Roles {
		// Remove all other roles
		err = s.GuildMemberRoleRemove(m.GuildID, m.User.ID, role)
		if err != nil {
			utils.SendToDevChannelDMs(s, fmt.Sprintf("Error removing role: %v", err), 1)
			return
		}
	}

	// Isolate the user
	err = s.GuildMemberRoleAdd(m.GuildID, m.User.ID, isolationRole)
	if err != nil {
		utils.SendToDevChannelDMs(s, fmt.Sprintf("Error giving isolation role: %v", err), 1)
	}
}

// HandleLeave processes member leave events
func (b *Bot) HandleLeave(s *discordgo.Session, m *discordgo.GuildMemberRemove) {
	// Nothing for now
	return
}
