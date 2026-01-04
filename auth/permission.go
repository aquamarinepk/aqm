package auth

import "strings"

type Permission string

func (p Permission) String() string {
	return string(p)
}

func (p Permission) Matches(required Permission) bool {
	if p == "*" {
		return true
	}
	if p == required {
		return true
	}

	pParts := strings.Split(string(p), ":")
	reqParts := strings.Split(string(required), ":")

	if len(pParts) != len(reqParts) {
		return false
	}

	for i := range pParts {
		if pParts[i] != "*" && pParts[i] != reqParts[i] {
			return false
		}
	}

	return true
}

func HasPermission(permissions []string, required string) bool {
	req := Permission(required)
	for _, perm := range permissions {
		if Permission(perm).Matches(req) {
			return true
		}
	}
	return false
}

func HasAnyPermission(permissions []string, required []string) bool {
	for _, req := range required {
		if HasPermission(permissions, req) {
			return true
		}
	}
	return false
}

func HasAllPermissions(permissions []string, required []string) bool {
	for _, req := range required {
		if !HasPermission(permissions, req) {
			return false
		}
	}
	return true
}
