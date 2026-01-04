package handler

import (
	"encoding/json"
	"net/http"

	"github.com/aquamarinepk/aqm/auth"
	"github.com/aquamarinepk/aqm/auth/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type AuthZHandler struct {
	roleStore  auth.RoleStore
	grantStore auth.GrantStore
}

func NewAuthZHandler(roleStore auth.RoleStore, grantStore auth.GrantStore) *AuthZHandler {
	return &AuthZHandler{
		roleStore:  roleStore,
		grantStore: grantStore,
	}
}

func (h *AuthZHandler) RegisterRoutes(r chi.Router) {
	r.Post("/roles", h.handleCreateRole)
	r.Get("/roles/{id}", h.handleGetRole)
	r.Get("/roles/name/{name}", h.handleGetRoleByName)
	r.Get("/roles", h.handleListRoles)
	r.Put("/roles/{id}", h.handleUpdateRole)
	r.Delete("/roles/{id}", h.handleDeleteRole)

	r.Post("/grants", h.handleAssignRole)
	r.Delete("/grants", h.handleRevokeRole)
	r.Get("/users/{user_id}/roles", h.handleGetUserRoles)
	r.Get("/users/{user_id}/grants", h.handleGetUserGrants)
	r.Get("/roles/{role_id}/grants", h.handleGetRoleGrants)

	r.Get("/users/{user_id}/permissions/{permission}", h.handleCheckPermission)
	r.Post("/users/{user_id}/check-any-permission", h.handleCheckAnyPermission)
	r.Post("/users/{user_id}/check-all-permissions", h.handleCheckAllPermissions)
	r.Get("/users/{user_id}/has-role/{role_name}", h.handleHasRole)
}

type CreateRoleRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
	CreatedBy   string   `json:"created_by"`
}

type RoleResponse struct {
	Role *auth.Role `json:"role"`
}

func (h *AuthZHandler) handleCreateRole(w http.ResponseWriter, r *http.Request) {
	var req CreateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	role, err := service.CreateRole(
		r.Context(),
		h.roleStore,
		req.Name,
		req.Description,
		req.Permissions,
		req.CreatedBy,
	)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, RoleResponse{Role: role})
}

func (h *AuthZHandler) handleGetRole(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	roleID, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ROLE_ID", "Invalid role ID format")
		return
	}

	role, err := service.GetRoleByID(r.Context(), h.roleStore, roleID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, RoleResponse{Role: role})
}

func (h *AuthZHandler) handleGetRoleByName(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	role, err := service.GetRoleByName(r.Context(), h.roleStore, name)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, RoleResponse{Role: role})
}

type ListRolesResponse struct {
	Roles []*auth.Role `json:"roles"`
}

func (h *AuthZHandler) handleListRoles(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")

	var roles []*auth.Role
	var err error

	if status != "" {
		roles, err = service.ListRolesByStatus(r.Context(), h.roleStore, auth.RoleStatus(status))
	} else {
		roles, err = service.ListRoles(r.Context(), h.roleStore)
	}

	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, ListRolesResponse{Roles: roles})
}

type UpdateRoleRequest struct {
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
	UpdatedBy   string   `json:"updated_by"`
}

func (h *AuthZHandler) handleUpdateRole(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	roleID, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ROLE_ID", "Invalid role ID format")
		return
	}

	var req UpdateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	role, err := service.GetRoleByID(r.Context(), h.roleStore, roleID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	role.Description = req.Description
	role.Permissions = req.Permissions

	if err := service.UpdateRole(r.Context(), h.roleStore, role, req.UpdatedBy); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, RoleResponse{Role: role})
}

func (h *AuthZHandler) handleDeleteRole(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	roleID, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ROLE_ID", "Invalid role ID format")
		return
	}

	if err := service.DeleteRole(r.Context(), h.roleStore, roleID); err != nil {
		handleServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type AssignRoleRequest struct {
	UserID     string `json:"user_id"`
	RoleID     string `json:"role_id"`
	AssignedBy string `json:"assigned_by"`
}

type GrantResponse struct {
	Grant *auth.Grant `json:"grant"`
}

func (h *AuthZHandler) handleAssignRole(w http.ResponseWriter, r *http.Request) {
	var req AssignRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_USER_ID", "Invalid user ID format")
		return
	}

	roleID, err := uuid.Parse(req.RoleID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ROLE_ID", "Invalid role ID format")
		return
	}

	grant, err := service.AssignRole(r.Context(), h.grantStore, userID, roleID, req.AssignedBy)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, GrantResponse{Grant: grant})
}

type RevokeRoleRequest struct {
	UserID string `json:"user_id"`
	RoleID string `json:"role_id"`
}

func (h *AuthZHandler) handleRevokeRole(w http.ResponseWriter, r *http.Request) {
	var req RevokeRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_USER_ID", "Invalid user ID format")
		return
	}

	roleID, err := uuid.Parse(req.RoleID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ROLE_ID", "Invalid role ID format")
		return
	}

	if err := service.RevokeRole(r.Context(), h.grantStore, userID, roleID); err != nil {
		handleServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type UserRolesResponse struct {
	Roles []*auth.Role `json:"roles"`
}

func (h *AuthZHandler) handleGetUserRoles(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_USER_ID", "Invalid user ID format")
		return
	}

	roles, err := service.GetUserRoles(r.Context(), h.grantStore, userID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, UserRolesResponse{Roles: roles})
}

type UserGrantsResponse struct {
	Grants []*auth.Grant `json:"grants"`
}

func (h *AuthZHandler) handleGetUserGrants(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_USER_ID", "Invalid user ID format")
		return
	}

	grants, err := service.GetUserGrants(r.Context(), h.grantStore, userID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, UserGrantsResponse{Grants: grants})
}

type RoleGrantsResponse struct {
	Grants []*auth.Grant `json:"grants"`
}

func (h *AuthZHandler) handleGetRoleGrants(w http.ResponseWriter, r *http.Request) {
	roleIDStr := chi.URLParam(r, "role_id")
	roleID, err := uuid.Parse(roleIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ROLE_ID", "Invalid role ID format")
		return
	}

	grants, err := service.GetRoleGrants(r.Context(), h.grantStore, roleID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, RoleGrantsResponse{Grants: grants})
}

type PermissionCheckResponse struct {
	HasPermission bool `json:"has_permission"`
}

func (h *AuthZHandler) handleCheckPermission(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_USER_ID", "Invalid user ID format")
		return
	}

	permission := chi.URLParam(r, "permission")

	hasPermission, err := service.CheckPermission(r.Context(), h.grantStore, userID, permission)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, PermissionCheckResponse{HasPermission: hasPermission})
}

type CheckAnyPermissionRequest struct {
	Permissions []string `json:"permissions"`
}

func (h *AuthZHandler) handleCheckAnyPermission(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_USER_ID", "Invalid user ID format")
		return
	}

	var req CheckAnyPermissionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	hasPermission, err := service.CheckAnyPermission(r.Context(), h.grantStore, userID, req.Permissions)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, PermissionCheckResponse{HasPermission: hasPermission})
}

type CheckAllPermissionsRequest struct {
	Permissions []string `json:"permissions"`
}

func (h *AuthZHandler) handleCheckAllPermissions(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_USER_ID", "Invalid user ID format")
		return
	}

	var req CheckAllPermissionsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	hasPermission, err := service.CheckAllPermissions(r.Context(), h.grantStore, userID, req.Permissions)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, PermissionCheckResponse{HasPermission: hasPermission})
}

type HasRoleResponse struct {
	HasRole bool `json:"has_role"`
}

func (h *AuthZHandler) handleHasRole(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_USER_ID", "Invalid user ID format")
		return
	}

	roleName := chi.URLParam(r, "role_name")

	hasRole, err := service.HasRole(r.Context(), h.grantStore, userID, roleName)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, HasRoleResponse{HasRole: hasRole})
}
