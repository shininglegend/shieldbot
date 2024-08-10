// internal/permissions/permissions.go

package permissions

import (
	"database/sql"
	"sync"

	"github.com/bwmarrin/discordgo"
)

type PermissionManager struct {
	db    *sql.DB
	cache struct {
		sync.RWMutex
		permissions map[string]map[string]string // guildID -> commandName -> roleID
	}
}

func (pm *PermissionManager) IsAdmin(s *discordgo.Session, guildID, userID string) (bool, error) {
	member, err := s.GuildMember(guildID, userID)
	if err != nil {
		return false, err
	}

	for _, roleID := range member.Roles {
		role, err := s.State.Role(guildID, roleID)
		if err != nil {
			return false, err
		}
		if role.Permissions&discordgo.PermissionAdministrator == discordgo.PermissionAdministrator {
			return true, nil
		}
	}

	return false, nil
}

func NewPermissionManager(db *sql.DB) *PermissionManager {
	pm := &PermissionManager{
		db: db,
	}
	pm.cache.permissions = make(map[string]map[string]string)
	return pm
}

func (pm *PermissionManager) loadPermissions() error {
	pm.cache.Lock()
	defer pm.cache.Unlock()

	rows, err := pm.db.Query("SELECT guild_id, command_name, role_id FROM command_permissions")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var guildID, commandName, roleID string
		if err := rows.Scan(&guildID, &commandName, &roleID); err != nil {
			return err
		}
		if _, ok := pm.cache.permissions[guildID]; !ok {
			pm.cache.permissions[guildID] = make(map[string]string)
		}
		pm.cache.permissions[guildID][commandName] = roleID
	}

	return rows.Err()
}

func (pm *PermissionManager) SetCommandPermission(guildID, commandName, roleID string) error {
	_, err := pm.db.Exec(`
		INSERT INTO command_permissions (guild_id, command_name, role_id)
		VALUES (?, ?, ?)
		ON CONFLICT(guild_id, command_name) DO UPDATE SET role_id = ?`,
		guildID, commandName, roleID, roleID)
	if err != nil {
		return err
	}

	// Update the cache
	pm.cache.Lock()
	defer pm.cache.Unlock()
	if _, ok := pm.cache.permissions[guildID]; !ok {
		pm.cache.permissions[guildID] = make(map[string]string)
	}
	pm.cache.permissions[guildID][commandName] = roleID

	return nil
}

func (pm *PermissionManager) CanUseCommand(guildID, userID, commandName string) (bool, error) {
	pm.cache.RLock()
	roleID, ok := pm.cache.permissions[guildID][commandName]
	pm.cache.RUnlock()

	if !ok {
		// If not in cache, check the database
		err := pm.db.QueryRow("SELECT role_id FROM command_permissions WHERE guild_id = ? AND command_name = ?", guildID, commandName).Scan(&roleID)
		if err != nil {
			if err == sql.ErrNoRows {
				return false, nil
			}
			return false, err
		}

		// Update cache
		pm.cache.Lock()
		if _, ok := pm.cache.permissions[guildID]; !ok {
			pm.cache.permissions[guildID] = make(map[string]string)
		}
		pm.cache.permissions[guildID][commandName] = roleID
		pm.cache.Unlock()
	}

	var hasRole bool
	err := pm.db.QueryRow("SELECT EXISTS(SELECT 1 FROM user_roles WHERE user_id = ? AND guild_id = ? AND roles LIKE ?)", userID, guildID, "%"+roleID+"%").Scan(&hasRole)
	if err != nil {
		return false, err
	}

	return hasRole, nil
}

func (pm *PermissionManager) SetIsolationRole(guildID, roleID string) error {
	_, err := pm.db.Exec(`
		INSERT INTO guild_settings (guild_id, setting_name, role_id)
		VALUES (?, 'isolation_role', ?)
		ON CONFLICT(guild_id, setting_name) DO UPDATE SET role_id = ?`,
		guildID, roleID, roleID)
	return err
}

func (pm *PermissionManager) GetIsolationRoleID(guildID string) (string, error) {
	var roleID string
	err := pm.db.QueryRow("SELECT role_id FROM guild_settings WHERE guild_id = ? AND setting_name = 'isolation_role'", guildID).Scan(&roleID)
	if err != nil {
		return "", err
	}
	return roleID, nil
}

func (pm *PermissionManager) SetupTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS command_permissions (
			guild_id TEXT,
			command_name TEXT,
			role_id TEXT,
			PRIMARY KEY (guild_id, command_name)
		)`,
		`CREATE TABLE IF NOT EXISTS guild_settings (
			guild_id TEXT,
			setting_name TEXT,
			role_id TEXT,
			PRIMARY KEY (guild_id, setting_name)
		)`,
	}

	for _, query := range queries {
		_, err := pm.db.Exec(query)
		if err != nil {
			return err
		}
	}

	return pm.loadPermissions()
}
