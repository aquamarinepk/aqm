package auth

type UserStatus string

const (
	UserStatusActive    UserStatus = "active"
	UserStatusSuspended UserStatus = "suspended"
	UserStatusPending   UserStatus = "pending"
	UserStatusDeleted   UserStatus = "deleted"
)

func (s UserStatus) String() string {
	return string(s)
}

func (s UserStatus) IsValid() bool {
	switch s {
	case UserStatusActive, UserStatusSuspended, UserStatusPending, UserStatusDeleted:
		return true
	default:
		return false
	}
}

type RoleStatus string

const (
	RoleStatusActive   RoleStatus = "active"
	RoleStatusInactive RoleStatus = "inactive"
)

func (s RoleStatus) String() string {
	return string(s)
}

func (s RoleStatus) IsValid() bool {
	switch s {
	case RoleStatusActive, RoleStatusInactive:
		return true
	default:
		return false
	}
}
