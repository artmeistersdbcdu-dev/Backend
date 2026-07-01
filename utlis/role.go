package utlis

import "github.com/Blue-Onion/ArtmeisterBackend/internal/database"

func GetRoleWeight(r database.UserRole) int {
	switch r {
	case database.UserRolePresident:
		return 7
	case database.UserRoleVicePresident:
		return 6
	case database.UserRoleGeneralSecretary:
		return 5
	case database.UserRoleLogistic:
		return 4
	case database.UserRoleSocialMediaHead:
		return 3
	case database.UserRoleContentHead:
		return 2
	case database.UserRoleCoreMember:
		return 1
	case database.UserRoleMember:
		return 0
	default:
		return -1
	}
}

func CanModerate(r database.UserRole) bool {
	return r == database.UserRolePresident || r == database.UserRoleVicePresident
}

func CanAssignRole(actor, target database.UserRole) bool {
	return GetRoleWeight(actor) > GetRoleWeight(target)
}

func IsValidUserRole(s string) bool {
	switch database.UserRole(s) {
	case database.UserRolePresident,
		database.UserRoleVicePresident,
		database.UserRoleGeneralSecretary,
		database.UserRoleLogistic,
		database.UserRoleSocialMediaHead,
		database.UserRoleContentHead,
		database.UserRoleCoreMember,
		database.UserRoleMember:
		return true
	default:
		return false
	}
}
