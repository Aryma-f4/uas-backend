package helpers

import (
	"context"

	"github.com/google/uuid"
)

// CheckPermission verifies if user has permission
func CheckPermission(c context.Context, userID uuid.UUID, requiredPermission string, permissionMap map[uuid.UUID][]string) bool {
	permissions, exists := permissionMap[userID]
	if !exists {
		return false
	}
	
	for _, p := range permissions {
		if p == requiredPermission {
			return true
		}
	}
	
	return false
}

// HasAnyPermission checks if user has any of the required permissions
func HasAnyPermission(userID uuid.UUID, requiredPermissions []string, permissionMap map[uuid.UUID][]string) bool {
	userPerms, exists := permissionMap[userID]
	if !exists {
		return false
	}
	
	for _, required := range requiredPermissions {
		for _, perm := range userPerms {
			if perm == required {
				return true
			}
		}
	}
	
	return false
}

// HasAllPermissions checks if user has all required permissions
func HasAllPermissions(userID uuid.UUID, requiredPermissions []string, permissionMap map[uuid.UUID][]string) bool {
	userPerms, exists := permissionMap[userID]
	if !exists {
		return false
	}
	
	for _, required := range requiredPermissions {
		found := false
		for _, perm := range userPerms {
			if perm == required {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	
	return true
}
