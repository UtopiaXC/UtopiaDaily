package service

import (
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/UtopiaXC/UtopiaDaily/internal/common"
	"github.com/UtopiaXC/UtopiaDaily/internal/event"
	"github.com/UtopiaXC/UtopiaDaily/internal/model"
	"github.com/UtopiaXC/UtopiaDaily/internal/repository"
	"github.com/UtopiaXC/UtopiaDaily/internal/security"
)

type UserService struct {
	userRepo *repository.UserRepository
	roleRepo *repository.RoleRepository
	bus      *event.Bus
}

func NewUserService(userRepo *repository.UserRepository, roleRepo *repository.RoleRepository, bus *event.Bus) *UserService {
	return &UserService{
		userRepo: userRepo,
		roleRepo: roleRepo,
		bus:      bus,
	}
}

func (s *UserService) ListUsers() ([]UserDTO, error) {
	users, err := s.userRepo.ListUsers()
	if err != nil {
		return nil, common.NewAPIError(http.StatusInternalServerError, common.ErrDatabaseError)
	}
	out := make([]UserDTO, 0, len(users))
	for i := range users {
		out = append(out, ToUserDTO(&users[i]))
	}
	return out, nil
}

func (s *UserService) CreateUser(operator string, username, nickname, email, password string, roleID uint) (UserDTO, error) {
	username = strings.TrimSpace(username)
	if !usernameRegex.MatchString(username) {
		return UserDTO{}, common.NewAPIError(http.StatusBadRequest, common.ErrAuthUsernameInvalid)
	}
	if len(password) < 8 {
		return UserDTO{}, common.NewAPIError(http.StatusBadRequest, common.ErrAuthPasswordTooShort)
	}
	email = strings.TrimSpace(email)
	var emailPtr *string
	if email != "" {
		if _, err := mail.ParseAddress(email); err != nil {
			return UserDTO{}, common.NewAPIError(http.StatusBadRequest, common.ErrAuthEmailInvalid)
		}
		if _, err := s.userRepo.FindByEmail(email); err == nil {
			return UserDTO{}, common.NewAPIError(http.StatusBadRequest, common.ErrAuthEmailTaken)
		}
		emailPtr = &email
	}
	if _, err := s.userRepo.FindByUsername(username); err == nil {
		return UserDTO{}, common.NewAPIError(http.StatusBadRequest, common.ErrAuthUsernameTaken)
	}
	role, err := s.roleRepo.FindByID(roleID)
	if err != nil {
		return UserDTO{}, common.NewAPIError(http.StatusBadRequest, common.ErrRoleNotFound)
	}

	hash, err := security.HashPassword(password)
	if err != nil {
		return UserDTO{}, common.NewAPIError(http.StatusInternalServerError, common.ErrInternalServerError)
	}
	if strings.TrimSpace(nickname) == "" {
		nickname = username
	}

	now := time.Now().UTC()
	u := &model.User{
		Username:     username,
		Nickname:     strings.TrimSpace(nickname),
		Email:        emailPtr,
		PasswordHash: hash,
		RoleID:       role.ID,
		Status:       model.UserStatusActive,
		TokenVersion: 1,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := s.userRepo.Create(u); err != nil {
		return UserDTO{}, common.SanitizeDBError(err)
	}
	fresh, err := s.userRepo.FindByID(u.ID)
	if err != nil {
		return UserDTO{}, common.NewAPIError(http.StatusInternalServerError, common.ErrDatabaseError)
	}

	if s.bus != nil {
		s.bus.Emit(model.EventLevelNormal, model.EventCategoryUser, "event.user.created", map[string]any{
			"username": fresh.Username,
			"operator": operator,
		}, nil, "user", fresh.Username)
	}
	return ToUserDTO(fresh), nil
}

func (s *UserService) UpdateUser(operator string, operatorID, targetID uint, nickname, email string, roleID uint, status string) (UserDTO, error) {
	target, err := s.userRepo.FindByID(targetID)
	if err != nil {
		return UserDTO{}, common.NewAPIError(http.StatusNotFound, common.ErrAuthUserNotFound)
	}
	newRole, err := s.roleRepo.FindByID(roleID)
	if err != nil {
		return UserDTO{}, common.NewAPIError(http.StatusBadRequest, common.ErrRoleNotFound)
	}

	if status == "" {
		status = target.Status
	}

	isCurrentAdmin := target.Role != nil && target.Role.Name == "admin" && target.Status == model.UserStatusActive
	willRemainActiveAdmin := newRole.Name == "admin" && status == model.UserStatusActive
	if isCurrentAdmin && !willRemainActiveAdmin {
		otherAdmins, err := s.userRepo.CountActiveAdmins(target.ID)
		if err == nil && otherAdmins == 0 {
			return UserDTO{}, common.NewAPIError(http.StatusBadRequest, common.ErrUserCannotDeactivateAdmin)
		}
	}
	if targetID == operatorID && status != model.UserStatusActive {
		return UserDTO{}, common.NewAPIError(http.StatusBadRequest, common.ErrUserCannotDeleteSelf)
	}

	updates := map[string]any{
		"nickname": strings.TrimSpace(nickname),
		"role_id":  roleID,
		"status":   status,
	}
	email = strings.TrimSpace(email)
	if email != "" {
		if _, err := mail.ParseAddress(email); err != nil {
			return UserDTO{}, common.NewAPIError(http.StatusBadRequest, common.ErrAuthEmailInvalid)
		}
		if existing, err := s.userRepo.FindByEmail(email); err == nil && existing.ID != targetID {
			return UserDTO{}, common.NewAPIError(http.StatusBadRequest, common.ErrAuthEmailTaken)
		}
		updates["email"] = &email
	} else {
		updates["email"] = nil
	}

	if err := s.userRepo.UpdateFields(targetID, updates); err != nil {
		return UserDTO{}, common.SanitizeDBError(err)
	}
	if target.RoleID != roleID || status != model.UserStatusActive {
		_ = s.userRepo.IncrementTokenVersionAndRevokeSessions(targetID)
	}

	fresh, err := s.userRepo.FindByID(targetID)
	if err != nil {
		return UserDTO{}, common.NewAPIError(http.StatusInternalServerError, common.ErrDatabaseError)
	}
	if s.bus != nil {
		s.bus.Emit(model.EventLevelNormal, model.EventCategoryUser, "event.user.updated", map[string]any{
			"username": fresh.Username,
			"operator": operator,
		}, nil, "user", fresh.Username)
	}
	return ToUserDTO(fresh), nil
}

func (s *UserService) ResetUserPassword(operator string, targetID uint, newPassword string) error {
	if len(newPassword) < 8 {
		return common.NewAPIError(http.StatusBadRequest, common.ErrAuthPasswordTooShort)
	}
	target, err := s.userRepo.FindByID(targetID)
	if err != nil {
		return common.NewAPIError(http.StatusNotFound, common.ErrAuthUserNotFound)
	}
	hash, err := security.HashPassword(newPassword)
	if err != nil {
		return common.NewAPIError(http.StatusInternalServerError, common.ErrInternalServerError)
	}
	if err := s.userRepo.UpdateFields(targetID, map[string]any{
		"password_hash": hash,
	}); err != nil {
		return common.NewAPIError(http.StatusInternalServerError, common.ErrDatabaseError)
	}
	_ = s.userRepo.IncrementTokenVersionAndRevokeSessions(targetID)
	if s.bus != nil {
		s.bus.Emit(model.EventLevelNormal, model.EventCategoryUser, "event.user.updated", map[string]any{
			"username": target.Username,
			"operator": operator,
		}, nil, "user", target.Username)
	}
	return nil
}

func (s *UserService) BanUser(operator string, operatorID, targetID uint) error {
	if operatorID == targetID {
		return common.NewAPIError(http.StatusBadRequest, common.ErrUserCannotDeleteSelf)
	}
	target, err := s.userRepo.FindByID(targetID)
	if err != nil {
		return common.NewAPIError(http.StatusNotFound, common.ErrAuthUserNotFound)
	}
	if target.Role != nil && target.Role.Name == "admin" && target.Status == model.UserStatusActive {
		otherAdmins, err := s.userRepo.CountActiveAdmins(targetID)
		if err == nil && otherAdmins == 0 {
			return common.NewAPIError(http.StatusBadRequest, common.ErrUserCannotDeactivateAdmin)
		}
	}
	if err := s.userRepo.UpdateFields(targetID, map[string]any{
		"status": model.UserStatusBanned,
	}); err != nil {
		return common.NewAPIError(http.StatusInternalServerError, common.ErrDatabaseError)
	}
	_ = s.userRepo.IncrementTokenVersionAndRevokeSessions(targetID)
	if s.bus != nil {
		s.bus.Emit(model.EventLevelWarning, model.EventCategoryUser, "event.user.banned", map[string]any{
			"username": target.Username,
			"operator": operator,
		}, nil, "user", target.Username)
	}
	return nil
}

func (s *UserService) UnfreezeUser(operator string, targetID uint) error {
	target, err := s.userRepo.FindByID(targetID)
	if err != nil {
		return common.NewAPIError(http.StatusNotFound, common.ErrAuthUserNotFound)
	}
	if err := s.userRepo.UpdateFields(targetID, map[string]any{
		"status":                model.UserStatusActive,
		"failed_login_attempts": 0,
		"lockout_until":         nil,
	}); err != nil {
		return common.NewAPIError(http.StatusInternalServerError, common.ErrDatabaseError)
	}
	if s.bus != nil {
		s.bus.Emit(model.EventLevelNormal, model.EventCategoryUser, "event.user.unfrozen", map[string]any{
			"username": target.Username,
			"operator": operator,
		}, nil, "user", target.Username)
	}
	return nil
}

func (s *UserService) DeleteUser(operator string, operatorID, targetID uint) error {
	if operatorID == targetID {
		return common.NewAPIError(http.StatusBadRequest, common.ErrUserCannotDeleteSelf)
	}
	target, err := s.userRepo.FindByID(targetID)
	if err != nil {
		return common.NewAPIError(http.StatusNotFound, common.ErrAuthUserNotFound)
	}
	if target.Role != nil && target.Role.Name == "admin" && target.Status == model.UserStatusActive {
		otherAdmins, err := s.userRepo.CountActiveAdmins(targetID)
		if err == nil && otherAdmins == 0 {
			return common.NewAPIError(http.StatusBadRequest, common.ErrUserCannotDeleteLastAdmin)
		}
	}
	_ = s.userRepo.IncrementTokenVersionAndRevokeSessions(targetID)
	if err := s.userRepo.Delete(targetID); err != nil {
		return common.SanitizeDBError(err)
	}
	if s.bus != nil {
		s.bus.Emit(model.EventLevelWarning, model.EventCategoryUser, "event.user.deleted", map[string]any{
			"username": target.Username,
			"operator": operator,
		}, nil, "user", target.Username)
	}
	return nil
}

func (s *UserService) ListRoles() ([]repository.RoleWithCount, error) {
	return s.roleRepo.ListRolesWithCounts()
}

func (s *UserService) ListPermissions() ([]model.Permission, error) {
	return s.roleRepo.ListPermissions()
}

func (s *UserService) CreateRole(name, description string, permCodes []string) (*model.Role, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, common.NewAPIError(http.StatusBadRequest, common.ErrValidationFailed)
	}
	if _, err := s.roleRepo.FindByName(name); err == nil {
		return nil, common.NewAPIError(http.StatusBadRequest, common.ErrRoleNameTaken)
	}
	perms, err := s.resolvePermissions(permCodes)
	if err != nil {
		return nil, err
	}
	role, err := s.roleRepo.CreateRole(name, strings.TrimSpace(description), perms)
	if err != nil {
		return nil, common.SanitizeDBError(err)
	}
	if s.bus != nil {
		s.bus.Emit(model.EventLevelNormal, model.EventCategoryUser, "event.role.created", map[string]any{
			"role": role.Name,
		}, nil, "role", role.Name)
	}
	return role, nil
}

func (s *UserService) UpdateRole(roleID uint, name, description string, permCodes []string) error {
	role, err := s.roleRepo.FindByID(roleID)
	if err != nil {
		return common.NewAPIError(http.StatusNotFound, common.ErrRoleNotFound)
	}
	if role.Name == "admin" {
		return common.NewAPIError(http.StatusForbidden, common.ErrRoleCannotEditAdmin)
	}
	name = strings.TrimSpace(name)
	if name == "" {
		name = role.Name
	}
	if role.IsSystem {
		name = role.Name
	}
	perms, err := s.resolvePermissions(permCodes)
	if err != nil {
		return err
	}
	if err := s.roleRepo.UpdateRole(roleID, name, strings.TrimSpace(description), perms); err != nil {
		return common.SanitizeDBError(err)
	}
	if s.bus != nil {
		s.bus.Emit(model.EventLevelNormal, model.EventCategoryUser, "event.role.updated", map[string]any{
			"role": name,
		}, nil, "role", name)
	}
	return nil
}

func (s *UserService) DeleteRole(roleID uint) error {
	role, err := s.roleRepo.FindByID(roleID)
	if err != nil {
		return common.NewAPIError(http.StatusNotFound, common.ErrRoleNotFound)
	}
	if role.IsSystem || role.Name == "admin" {
		return common.NewAPIError(http.StatusForbidden, common.ErrRoleCannotDeleteSystem)
	}
	cnt, err := s.roleRepo.CountUsersInRole(roleID)
	if err == nil && cnt > 0 {
		return common.NewAPIError(http.StatusBadRequest, common.ErrRoleHasUsers)
	}
	if err := s.roleRepo.DeleteRole(roleID); err != nil {
		return common.SanitizeDBError(err)
	}
	if s.bus != nil {
		s.bus.Emit(model.EventLevelNormal, model.EventCategoryUser, "event.role.deleted", map[string]any{
			"role": role.Name,
		}, nil, "role", role.Name)
	}
	return nil
}

func (s *UserService) resolvePermissions(codes []string) ([]model.Permission, error) {
	if len(codes) == 0 {
		return nil, nil
	}
	perms, err := s.roleRepo.FindPermissionsByCodes(codes)
	if err != nil {
		return nil, common.NewAPIError(http.StatusInternalServerError, common.ErrDatabaseError)
	}
	found := make(map[string]bool, len(perms))
	for _, p := range perms {
		found[p.Code] = true
	}
	for _, code := range codes {
		if !found[code] {
			return nil, common.NewAPIError(http.StatusBadRequest, common.ErrRolePermissionInvalid)
		}
	}
	return perms, nil
}
