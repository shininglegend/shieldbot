// internal/permissions/permissions.go

package permissions

import (
	"database/sql"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/shininglegend/shieldbot/pkg/utils"
)

type PermissionManager struct {
	db    *sql.DB
	cache struct {
		sync.RWMutex
		permissions map[string]map[string][]string // guildID -> commandName -> []roleID
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

	pm.cache.permissions = make(map[string]map[string][]string)
	return pm
}

func (pm *PermissionManager) loadPermissions() error {
	pm.cache.Lock()
	defer pm.cache.Unlock()

	// Clear existing cache
	pm.cache.permissions = make(map[string]map[string][]string)

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
			pm.cache.permissions[guildID] = make(map[string][]string)
		}

		pm.cache.permissions[guildID][commandName] = append(pm.cache.permissions[guildID][commandName], roleID)
	}

	return rows.Err()
}

func (pm *PermissionManager) GetCommandPermissions(guildID string) (map[string][]string, error) {
	pm.cache.RLock()
	cachedPerms, ok := pm.cache.permissions[guildID]
	pm.cache.RUnlock()

	if ok {
		return cachedPerms, nil
	}

	rows, err := pm.db.Query("SELECT command_name, role_id FROM command_permissions WHERE guild_id = ?", guildID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	perms := make(map[string][]string)
	for rows.Next() {
		var commandName, roleID string
		if err := rows.Scan(&commandName, &roleID); err != nil {
			return nil, err
		}
		perms[commandName] = append(perms[commandName], roleID)
	}

	// Update cache
	pm.cache.Lock()
	pm.cache.permissions[guildID] = perms
	pm.cache.Unlock()

	return perms, nil
}

// Updated SetCommandPermission function
func (pm *PermissionManager) SetCommandPermission(guildID, commandName, roleID string) error {
	_, err := pm.db.Exec(`
        INSERT OR IGNORE INTO command_permissions (guild_id, command_name, role_id)
        VALUES (?, ?, ?)`,
		guildID, commandName, roleID)
	if err != nil {
		return err
	}

	// Update the cache
	pm.cache.Lock()
	defer pm.cache.Unlock()
	if _, ok := pm.cache.permissions[guildID]; !ok {
		pm.cache.permissions[guildID] = make(map[string][]string)
	}
	if _, ok := pm.cache.permissions[guildID][commandName]; !ok {
		pm.cache.permissions[guildID][commandName] = []string{}
	}
	if !utils.Contains(pm.cache.permissions[guildID][commandName], roleID) {
		pm.cache.permissions[guildID][commandName] = append(pm.cache.permissions[guildID][commandName], roleID)
	}

	return nil
}

// Updated RemoveCommandPermission function
func (pm *PermissionManager) RemoveCommandPermission(guildID, commandName, roleID string) error {
	_, err := pm.db.Exec(`
        DELETE FROM command_permissions
        WHERE guild_id = ? AND command_name = ? AND role_id = ?`,
		guildID, commandName, roleID)
	if err != nil {
		return err
	}

	// Update the cache
	pm.cache.Lock()
	defer pm.cache.Unlock()
	if guildPerms, ok := pm.cache.permissions[guildID]; ok {
		if roles, ok := guildPerms[commandName]; ok {
			guildPerms[commandName] = utils.Remove(roles, roleID)
		}
	}

	return nil
}

// Updated CanUseCommand function
func (pm *PermissionManager) CanUseCommand(s *discordgo.Session, guildID, userID, commandName string) (bool, error) {
	pm.cache.RLock()
	roleIDs, ok := pm.cache.permissions[guildID][commandName]
	pm.cache.RUnlock()

	if !ok {
		// If not in cache, check the database
		rows, err := pm.db.Query("SELECT role_id FROM command_permissions WHERE guild_id = ? AND command_name = ?", guildID, commandName)
		if err != nil {
			return false, err
		}
		defer rows.Close()

		roleIDs = []string{}
		for rows.Next() {
			var roleID string
			if err := rows.Scan(&roleID); err != nil {
				return false, err
			}
			roleIDs = append(roleIDs, roleID)
		}

		// Update cache
		pm.cache.Lock()
		if _, ok := pm.cache.permissions[guildID]; !ok {
			pm.cache.permissions[guildID] = make(map[string][]string)
		}
		pm.cache.permissions[guildID][commandName] = roleIDs
		pm.cache.Unlock()
	}

	if len(roleIDs) == 0 {
		return false, nil
	}

	// Check if the user has any of the required roles
	member, err := s.GuildMember(guildID, userID)
	if err != nil {
		return false, err
	}

	// Very slow function :sad:
	for _, roleID := range roleIDs {
		for _, memberRoleID := range member.Roles {
			if memberRoleID == roleID {
				return true, nil
			}
		}
	}

	return false, nil
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
			PRIMARY KEY (guild_id, command_name, role_id)
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
