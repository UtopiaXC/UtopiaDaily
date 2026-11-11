package common

import (
	"errors"
	"net/http"
	"strings"

	"gorm.io/gorm"
)

const (
	ErrAuthInvalidCredentials   = "AUTH_INVALID_CREDENTIALS"
	ErrAuthUserNotFound         = "AUTH_USER_NOT_FOUND"
	ErrAuthUserBanned           = "AUTH_USER_BANNED"
	ErrAuthUserFrozen           = "AUTH_USER_FROZEN"
	ErrAuthUserInactive         = "AUTH_USER_INACTIVE"
	ErrAuthCaptchaRequired      = "AUTH_CAPTCHA_REQUIRED"
	ErrAuthCaptchaFailed        = "AUTH_CAPTCHA_FAILED"
	ErrAuthUnauthorized         = "AUTH_UNAUTHORIZED"
	ErrAuthTokenExpired         = "AUTH_TOKEN_EXPIRED"
	ErrAuthForbidden            = "AUTH_FORBIDDEN"
	ErrAuthPasswordTooShort     = "AUTH_PASSWORD_TOO_SHORT"
	ErrAuthPasswordMismatch     = "AUTH_PASSWORD_MISMATCH"
	ErrAuthOldPasswordWrong     = "AUTH_OLD_PASSWORD_WRONG"
	ErrAuthUsernameTaken        = "AUTH_USERNAME_TAKEN"
	ErrAuthUsernameInvalid      = "AUTH_USERNAME_INVALID"
	ErrAuthEmailTaken           = "AUTH_EMAIL_TAKEN"
	ErrAuthEmailInvalid         = "AUTH_EMAIL_INVALID"
	ErrAuthEmailNotConfigured   = "AUTH_EMAIL_NOT_CONFIGURED"
	ErrRegistrationDisabled     = "REGISTRATION_DISABLED"
	ErrRegistrationSMTPRequired = "REGISTRATION_SMTP_REQUIRED"
	ErrSMTPNotConfigured        = "SMTP_NOT_CONFIGURED"
	ErrSMTPTestFailed           = "SMTP_TEST_FAILED"
	ErrResetTokenInvalid        = "RESET_TOKEN_INVALID"
	ErrActivationTokenInvalid   = "ACTIVATION_TOKEN_INVALID"
	ErrSetupAlreadyCompleted    = "SETUP_ALREADY_COMPLETED"
	ErrUserEmailRequired        = "USER_EMAIL_REQUIRED"
	ErrCSRFBlocked              = "CSRF_BLOCKED"

	ErrUserCannotDeleteSelf      = "USER_CANNOT_DELETE_SELF"
	ErrUserCannotDeleteLastAdmin = "USER_CANNOT_DELETE_LAST_ADMIN"
	ErrUserCannotDeactivateAdmin = "USER_CANNOT_DEACTIVATE_LAST_ADMIN"
	ErrRoleNotFound              = "ROLE_NOT_FOUND"
	ErrRoleNameTaken             = "ROLE_NAME_TAKEN"
	ErrRoleHasUsers              = "ROLE_HAS_USERS"
	ErrRoleCannotDeleteSystem    = "ROLE_CANNOT_DELETE_SYSTEM"
	ErrRoleCannotEditAdmin       = "ROLE_CANNOT_EDIT_ADMIN"
	ErrRoleInvalidDefault        = "ROLE_INVALID_DEFAULT"
	ErrRolePermissionInvalid     = "ROLE_PERMISSION_INVALID"

	ErrModuleNotFound         = "MODULE_NOT_FOUND"
	ErrModuleAlreadyEnabled   = "MODULE_ALREADY_ENABLED"
	ErrModuleNotEnabled       = "MODULE_NOT_ENABLED"
	ErrModuleEnableFailed     = "MODULE_ENABLE_FAILED"
	ErrModuleTestFailed       = "MODULE_TEST_FAILED"
	ErrModuleExecutionFailed  = "MODULE_EXECUTION_FAILED"
	ErrPluginDependencyFailed = "PLUGIN_DEPENDENCY_FAILED"
	ErrScheduleTaskNotFound   = "SCHEDULE_TASK_NOT_FOUND"
	ErrScheduleCronInvalid    = "SCHEDULE_CRON_INVALID"
	ErrConfigNotFound         = "CONFIG_NOT_FOUND"
	ErrConfigNotEditable      = "CONFIG_NOT_EDITABLE"
	ErrConfigValidationFailed = "CONFIG_VALIDATION_FAILED"
	ErrSubscriptionNotFound   = "SUBSCRIPTION_NOT_FOUND"
	ErrPushDeliveryFailed     = "PUSH_DELIVERY_FAILED"

	ErrValidationFailed    = "VALIDATION_FAILED"
	ErrRecordNotFound      = "RECORD_NOT_FOUND"
	ErrDuplicateEntry      = "DUPLICATE_ENTRY"
	ErrConstraintViolation = "CONSTRAINT_VIOLATION"
	ErrDatabaseError       = "DATABASE_ERROR"
	ErrInternalServerError = "INTERNAL_SERVER_ERROR"
	ErrOperationFailed     = "OPERATION_FAILED"
)

type APIError struct {
	HTTPStatus int            `json:"-"`
	Code       string         `json:"code"`
	Params     map[string]any `json:"params,omitempty"`
}

func (e *APIError) Error() string {
	return e.Code
}

func NewAPIError(status int, code string, params ...map[string]any) *APIError {
	var p map[string]any
	if len(params) > 0 {
		p = params[0]
	}
	return &APIError{
		HTTPStatus: status,
		Code:       code,
		Params:     p,
	}
}

func SanitizeDBError(err error) *APIError {
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return NewAPIError(http.StatusNotFound, ErrRecordNotFound)
	}
	msg := err.Error()
	switch {
	case strings.Contains(msg, "users.email") || strings.Contains(msg, "idx_users_email"):
		return NewAPIError(http.StatusBadRequest, ErrAuthEmailTaken)
	case strings.Contains(msg, "users.username") || strings.Contains(msg, "idx_users_username"):
		return NewAPIError(http.StatusBadRequest, ErrAuthUsernameTaken)
	case strings.Contains(msg, "roles.name") || strings.Contains(msg, "idx_roles_name"):
		return NewAPIError(http.StatusBadRequest, ErrRoleNameTaken)
	case strings.Contains(msg, "UNIQUE constraint failed") || strings.Contains(msg, "Duplicate entry"):
		return NewAPIError(http.StatusBadRequest, ErrDuplicateEntry)
	case strings.Contains(msg, "FOREIGN KEY constraint failed") || strings.Contains(msg, "foreign key constraint"):
		return NewAPIError(http.StatusBadRequest, ErrConstraintViolation)
	default:
		return NewAPIError(http.StatusInternalServerError, ErrDatabaseError)
	}
}
