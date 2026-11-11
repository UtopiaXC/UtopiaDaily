package service

import (
	"errors"
	"net/http"
	"net/mail"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/UtopiaXC/UtopiaDaily/internal/common"
	"github.com/UtopiaXC/UtopiaDaily/internal/config"
	"github.com/UtopiaXC/UtopiaDaily/internal/event"
	"github.com/UtopiaXC/UtopiaDaily/internal/model"
	"github.com/UtopiaXC/UtopiaDaily/internal/pkg/mailer"
	"github.com/UtopiaXC/UtopiaDaily/internal/repository"
	"github.com/UtopiaXC/UtopiaDaily/internal/security"
	"gorm.io/gorm"
)

var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]{3,32}$`)

type UserDTO struct {
	ID          uint      `json:"id"`
	Username    string    `json:"username"`
	Nickname    string    `json:"nickname"`
	Email       string    `json:"email"`
	RoleID      uint      `json:"role_id"`
	RoleName    string    `json:"role_name"`
	IsAdmin     bool      `json:"is_admin"`
	Permissions []string  `json:"permissions"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

func ToUserDTO(u *model.User) UserDTO {
	dto := UserDTO{
		ID:        u.ID,
		Username:  u.Username,
		Nickname:  u.Nickname,
		RoleID:    u.RoleID,
		Status:    u.Status,
		CreatedAt: u.CreatedAt,
	}
	if u.Email != nil {
		dto.Email = *u.Email
	}
	if u.Role != nil {
		dto.RoleName = u.Role.Name
		dto.IsAdmin = u.Role.Name == "admin"
		perms := make([]string, 0, len(u.Role.Permissions))
		for _, p := range u.Role.Permissions {
			perms = append(perms, p.Code)
		}
		dto.Permissions = perms
	}
	if dto.Permissions == nil {
		dto.Permissions = []string{}
	}
	return dto
}

type LoginPolicyDTO struct {
	CaptchaRequired   bool `json:"captcha_required"`
	AllowRegistration bool `json:"allow_registration"`
	SMTPEnabled       bool `json:"smtp_enabled"`
}

type SetupInput struct {
	SiteName        string
	SiteDescription string
	BaseURL         string
	DefaultLocale   string
	SMTPEnabled     bool
	SMTPHost        string
	SMTPPort        int
	SMTPEncryption  string
	SMTPUser        string
	SMTPPassword    string
	SMTPFrom        string
	SMTPFromName    string
	Username        string
	Nickname        string
	Email           string
	Password        string
}

type AuthResponse struct {
	AccessToken  string  `json:"token"`
	RefreshToken string  `json:"refresh_token"`
	ExpiresIn    int64   `json:"expires_in"`
	User         UserDTO `json:"user"`
}

type AuthService struct {
	userRepo *repository.UserRepository
	roleRepo *repository.RoleRepository
	cfgRepo  *repository.ConfigRepository
	jwtMgr   *security.JWTManager
	captcha  security.CaptchaProvider
	attempts *security.LoginAttemptsTracker
	bus      *event.Bus
	baseCfg  *config.Config
}

func NewAuthService(
	userRepo *repository.UserRepository,
	roleRepo *repository.RoleRepository,
	cfgRepo *repository.ConfigRepository,
	jwtMgr *security.JWTManager,
	captcha security.CaptchaProvider,
	attempts *security.LoginAttemptsTracker,
	bus *event.Bus,
	baseCfg *config.Config,
) *AuthService {
	return &AuthService{
		userRepo: userRepo,
		roleRepo: roleRepo,
		cfgRepo:  cfgRepo,
		jwtMgr:   jwtMgr,
		captcha:  captcha,
		attempts: attempts,
		bus:      bus,
		baseCfg:  baseCfg,
	}
}

func (s *AuthService) IsSetupRequired() (bool, error) {
	hasAdmin, err := s.userRepo.HasAdmin()
	if err != nil {
		return false, err
	}
	return !hasAdmin, nil
}

func (s *AuthService) GetLoginPolicy() LoginPolicyDTO {
	smtpEnabled := s.getStringConfig("smtp_enabled", "false") == "true"
	allowReg := s.getStringConfig("allow_registration", "false") == "true" && smtpEnabled
	return LoginPolicyDTO{
		CaptchaRequired:   true,
		AllowRegistration: allowReg,
		SMTPEnabled:       smtpEnabled,
	}
}

func (s *AuthService) SetupAdmin(input SetupInput, ip, ua string) (*AuthResponse, error) {
	hasAdmin, err := s.userRepo.HasAdmin()
	if err != nil {
		return nil, common.NewAPIError(http.StatusInternalServerError, common.ErrDatabaseError)
	}
	if hasAdmin {
		return nil, common.NewAPIError(http.StatusConflict, common.ErrSetupAlreadyCompleted)
	}

	username := strings.TrimSpace(input.Username)
	if !usernameRegex.MatchString(username) {
		return nil, common.NewAPIError(http.StatusBadRequest, common.ErrAuthUsernameInvalid)
	}
	if len(input.Password) < 8 {
		return nil, common.NewAPIError(http.StatusBadRequest, common.ErrAuthPasswordTooShort)
	}
	email := strings.TrimSpace(input.Email)
	var emailPtr *string
	if email != "" {
		if _, err := mail.ParseAddress(email); err != nil {
			return nil, common.NewAPIError(http.StatusBadRequest, common.ErrAuthEmailInvalid)
		}
		emailPtr = &email
	}

	adminRole, err := s.roleRepo.FindByName("admin")
	if err != nil {
		return nil, common.NewAPIError(http.StatusInternalServerError, common.ErrRoleNotFound)
	}

	hash, err := security.HashPassword(input.Password)
	if err != nil {
		return nil, common.NewAPIError(http.StatusInternalServerError, common.ErrInternalServerError)
	}
	nickname := strings.TrimSpace(input.Nickname)
	if nickname == "" {
		nickname = username
	}

	cfgUpdates := make(map[string]string)
	if v := strings.TrimSpace(input.SiteName); v != "" {
		cfgUpdates["server_name"] = v
		cfgUpdates["newspaper_title"] = v
	}
	if v := strings.TrimSpace(input.SiteDescription); v != "" {
		cfgUpdates["site_description"] = v
	}
	if v := strings.TrimRight(strings.TrimSpace(input.BaseURL), "/"); v != "" {
		cfgUpdates["base_url"] = v
	}
	if v := strings.TrimSpace(input.DefaultLocale); v != "" {
		cfgUpdates["default_locale"] = v
	}
	if input.SMTPEnabled {
		cfgUpdates["smtp_enabled"] = "true"
		if v := strings.TrimSpace(input.SMTPHost); v != "" {
			cfgUpdates["smtp_host"] = v
		}
		if input.SMTPPort > 0 {
			cfgUpdates["smtp_port"] = strconv.Itoa(input.SMTPPort)
		}
		if v := strings.TrimSpace(input.SMTPEncryption); v != "" {
			cfgUpdates["smtp_encryption"] = v
		}
		if v := strings.TrimSpace(input.SMTPUser); v != "" {
			cfgUpdates["smtp_user"] = v
		}
		if input.SMTPPassword != "" {
			cfgUpdates["smtp_password"] = input.SMTPPassword
		}
		if v := strings.TrimSpace(input.SMTPFrom); v != "" {
			cfgUpdates["smtp_from"] = v
		}
		if v := strings.TrimSpace(input.SMTPFromName); v != "" {
			cfgUpdates["smtp_from_name"] = v
		}
	} else {
		cfgUpdates["smtp_enabled"] = "false"
	}
	if len(cfgUpdates) > 0 {
		if err := s.cfgRepo.SetMany(cfgUpdates); err != nil {
			return nil, common.NewAPIError(http.StatusInternalServerError, common.ErrDatabaseError)
		}
	}

	now := time.Now().UTC()
	user := &model.User{
		Username:     username,
		Nickname:     nickname,
		Email:        emailPtr,
		PasswordHash: hash,
		RoleID:       adminRole.ID,
		Status:       model.UserStatusActive,
		TokenVersion: 1,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := s.userRepo.Create(user); err != nil {
		return nil, common.SanitizeDBError(err)
	}

	created, err := s.userRepo.FindByID(user.ID)
	if err != nil {
		return nil, common.NewAPIError(http.StatusInternalServerError, common.ErrDatabaseError)
	}

	accessTTL, refreshTTL := s.getTokenTTLs()
	pair, err := s.jwtMgr.IssueTokens(created, ip, ua, true, accessTTL, refreshTTL)
	if err != nil {
		return nil, common.NewAPIError(http.StatusInternalServerError, common.ErrInternalServerError)
	}

	if s.bus != nil {
		s.bus.Emit(model.EventLevelNormal, model.EventCategorySystem, "event.system.setup_completed", map[string]any{
			"username": created.Username,
		}, nil, "user", created.Username)
	}

	return &AuthResponse{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		ExpiresIn:    pair.ExpiresIn,
		User:         ToUserDTO(created),
	}, nil
}

func (s *AuthService) GenerateCaptcha() (*security.Challenge, error) {
	difficulty := s.getIntConfig("pow_difficulty", 50000)
	if difficulty <= 0 {
		difficulty = 50000
	}
	return s.captcha.Generate(int64(difficulty))
}

func (s *AuthService) Login(account, password, powPayload string, rememberMe bool, ip, ua string) (*AuthResponse, error) {
	account = strings.TrimSpace(account)
	if account == "" || password == "" {
		return nil, common.NewAPIError(http.StatusBadRequest, common.ErrAuthInvalidCredentials)
	}

	if strings.TrimSpace(powPayload) == "" {
		return nil, common.NewAPIError(http.StatusBadRequest, common.ErrAuthCaptchaRequired)
	}
	if !s.captcha.Verify(powPayload) {
		return nil, common.NewAPIError(http.StatusBadRequest, common.ErrAuthCaptchaFailed)
	}

	lockKey := strings.ToLower(account) + "|" + ip
	if s.attempts.IsLocked(lockKey) {
		return nil, common.NewAPIError(http.StatusTooManyRequests, common.ErrAuthUserFrozen)
	}
	if delay := s.attempts.GetDelay(lockKey); delay > 0 {
		time.Sleep(delay)
	}

	maxAttempts := s.getIntConfig("max_login_attempts", 5)
	lockMinutes := s.baseCfg.Security.LockoutMinutes
	if lockMinutes <= 0 {
		lockMinutes = 15
	}
	lockDuration := time.Duration(lockMinutes) * time.Minute

	user, err := s.userRepo.FindByAccount(account)
	if err != nil {
		_, _ = s.attempts.RecordFailure(lockKey, maxAttempts, lockDuration)
		if s.bus != nil {
			s.bus.Emit(model.EventLevelWarning, model.EventCategorySecurity, "event.auth.login_failed", map[string]any{
				"account": account,
			}, map[string]any{"ip": ip}, "auth", account)
		}
		return nil, common.NewAPIError(http.StatusUnauthorized, common.ErrAuthInvalidCredentials)
	}

	switch user.Status {
	case model.UserStatusBanned:
		return nil, common.NewAPIError(http.StatusForbidden, common.ErrAuthUserBanned)
	case model.UserStatusInactive:
		return nil, common.NewAPIError(http.StatusForbidden, common.ErrAuthUserInactive)
	case model.UserStatusDisabled:
		return nil, common.NewAPIError(http.StatusForbidden, common.ErrAuthUserBanned)
	case model.UserStatusFrozen:
		if user.LockoutUntil != nil && time.Now().UTC().Before(*user.LockoutUntil) {
			return nil, common.NewAPIError(http.StatusTooManyRequests, common.ErrAuthUserFrozen)
		}
		_ = s.userRepo.UpdateFields(user.ID, map[string]any{
			"status":                model.UserStatusActive,
			"failed_login_attempts": 0,
			"lockout_until":         nil,
		})
		user.Status = model.UserStatusActive
	}

	if user.LockoutUntil != nil && time.Now().UTC().Before(*user.LockoutUntil) {
		return nil, common.NewAPIError(http.StatusTooManyRequests, common.ErrAuthUserFrozen)
	}

	ok, verErr := security.VerifyPassword(password, user.PasswordHash)
	if verErr != nil || !ok {
		_, locked := s.attempts.RecordFailure(lockKey, maxAttempts, lockDuration)
		newFailed := user.FailedLoginAttempts + 1
		updates := map[string]any{
			"failed_login_attempts": newFailed,
		}
		if locked || (maxAttempts > 0 && newFailed >= maxAttempts) {
			until := time.Now().UTC().Add(lockDuration)
			updates["lockout_until"] = &until
			updates["status"] = model.UserStatusFrozen
		}
		_ = s.userRepo.UpdateFields(user.ID, updates)

		if s.bus != nil {
			s.bus.Emit(model.EventLevelWarning, model.EventCategorySecurity, "event.auth.login_failed", map[string]any{
				"account": user.Username,
			}, map[string]any{"ip": ip}, "user", user.Username)
		}
		return nil, common.NewAPIError(http.StatusUnauthorized, common.ErrAuthInvalidCredentials)
	}

	s.attempts.RecordSuccess(lockKey)
	if user.FailedLoginAttempts > 0 || user.LockoutUntil != nil {
		_ = s.userRepo.UpdateFields(user.ID, map[string]any{
			"failed_login_attempts": 0,
			"lockout_until":         nil,
		})
	}

	accessTTL, refreshTTL := s.getTokenTTLs()
	if !rememberMe {
		refreshTTL = 24 * time.Hour
	}
	pair, err := s.jwtMgr.IssueTokens(user, ip, ua, rememberMe, accessTTL, refreshTTL)
	if err != nil {
		return nil, common.NewAPIError(http.StatusInternalServerError, common.ErrInternalServerError)
	}

	if s.bus != nil {
		s.bus.Emit(model.EventLevelNormal, model.EventCategoryUser, "event.auth.login_success", map[string]any{
			"username": user.Username,
		}, map[string]any{"ip": ip}, "user", user.Username)
	}

	return &AuthResponse{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		ExpiresIn:    pair.ExpiresIn,
		User:         ToUserDTO(user),
	}, nil
}

func (s *AuthService) RefreshToken(rawRefresh, ip, ua string) (*AuthResponse, error) {
	if strings.TrimSpace(rawRefresh) == "" {
		return nil, common.NewAPIError(http.StatusUnauthorized, common.ErrAuthUnauthorized)
	}
	accessTTL, refreshTTL := s.getTokenTTLs()
	pair, user, err := s.jwtMgr.RefreshTokens(rawRefresh, ip, ua, accessTTL, refreshTTL)
	if err != nil {
		return nil, common.NewAPIError(http.StatusUnauthorized, common.ErrAuthUnauthorized)
	}
	return &AuthResponse{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		ExpiresIn:    pair.ExpiresIn,
		User:         ToUserDTO(user),
	}, nil
}

func (s *AuthService) Logout(sessionID, rawRefresh string) error {
	if sessionID != "" {
		return s.jwtMgr.RevokeSession(sessionID)
	}
	if rawRefresh != "" {
		return s.jwtMgr.RevokeByRefreshToken(rawRefresh)
	}
	return nil
}

func (s *AuthService) Register(username, nickname, email, password, powPayload string) error {
	if s.getStringConfig("allow_registration", "false") != "true" {
		return common.NewAPIError(http.StatusForbidden, common.ErrRegistrationDisabled)
	}
	if s.getStringConfig("smtp_enabled", "false") != "true" {
		return common.NewAPIError(http.StatusBadRequest, common.ErrRegistrationSMTPRequired)
	}
	if strings.TrimSpace(powPayload) == "" {
		return common.NewAPIError(http.StatusBadRequest, common.ErrAuthCaptchaRequired)
	}
	if !s.captcha.Verify(powPayload) {
		return common.NewAPIError(http.StatusBadRequest, common.ErrAuthCaptchaFailed)
	}

	username = strings.TrimSpace(username)
	nickname = strings.TrimSpace(nickname)
	if nickname == "" {
		nickname = username
	}
	email = strings.TrimSpace(email)
	if !usernameRegex.MatchString(username) {
		return common.NewAPIError(http.StatusBadRequest, common.ErrAuthUsernameInvalid)
	}
	if email == "" {
		return common.NewAPIError(http.StatusBadRequest, common.ErrUserEmailRequired)
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return common.NewAPIError(http.StatusBadRequest, common.ErrAuthEmailInvalid)
	}
	if len(password) < 8 {
		return common.NewAPIError(http.StatusBadRequest, common.ErrAuthPasswordTooShort)
	}

	if _, err := s.userRepo.FindByUsername(username); err == nil {
		return common.NewAPIError(http.StatusBadRequest, common.ErrAuthUsernameTaken)
	}
	if _, err := s.userRepo.FindByEmail(email); err == nil {
		return common.NewAPIError(http.StatusBadRequest, common.ErrAuthEmailTaken)
	}

	defaultRoleID := uint(s.getIntConfig("default_register_role_id", 2))
	role, err := s.roleRepo.FindByID(defaultRoleID)
	if err != nil {
		role, err = s.roleRepo.FindByName("user")
		if err != nil {
			return common.NewAPIError(http.StatusInternalServerError, common.ErrRoleInvalidDefault)
		}
	}

	hash, err := security.HashPassword(password)
	if err != nil {
		return common.NewAPIError(http.StatusInternalServerError, common.ErrInternalServerError)
	}

	now := time.Now().UTC()
	user := &model.User{
		Username:     username,
		Nickname:     nickname,
		Email:        &email,
		PasswordHash: hash,
		RoleID:       role.ID,
		Status:       model.UserStatusInactive,
		TokenVersion: 1,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := s.userRepo.Create(user); err != nil {
		return common.SanitizeDBError(err)
	}

	rawToken, err := security.GenerateRandomToken(32)
	if err != nil {
		_ = s.userRepo.Delete(user.ID)
		return common.NewAPIError(http.StatusInternalServerError, common.ErrInternalServerError)
	}
	vToken := &model.VerificationToken{
		UserID:    user.ID,
		TokenHash: security.HashToken(rawToken),
		Purpose:   "activation",
		ExpiresAt: now.Add(24 * time.Hour),
		CreatedAt: now,
	}
	if err := s.userRepo.CreateVerificationToken(vToken); err != nil {
		_ = s.userRepo.Delete(user.ID)
		return common.NewAPIError(http.StatusInternalServerError, common.ErrDatabaseError)
	}

	baseURL := strings.TrimRight(s.getStringConfig("base_url", s.baseCfg.Site.BaseURL), "/")
	siteName := s.getStringConfig("server_name", s.baseCfg.Site.Name)
	link := baseURL + "/dashboard?mode=activate&token=" + rawToken
	subj, htmlBody, textBody := mailer.RenderActivationEmail(siteName, username, link)
	if err := s.buildMailer().SendHTML(email, subj, htmlBody, textBody); err != nil {
		_ = s.userRepo.Delete(user.ID)
		return common.NewAPIError(http.StatusBadGateway, common.ErrSMTPTestFailed)
	}

	if s.bus != nil {
		s.bus.Emit(model.EventLevelNormal, model.EventCategoryUser, "event.auth.user_registered", map[string]any{
			"username": username,
		}, nil, "user", username)
	}
	return nil
}

func (s *AuthService) ActivateAccount(rawToken string) error {
	rawToken = strings.TrimSpace(rawToken)
	if rawToken == "" {
		return common.NewAPIError(http.StatusBadRequest, common.ErrActivationTokenInvalid)
	}
	tok, err := s.userRepo.FindValidVerificationToken("activation", security.HashToken(rawToken))
	if err != nil {
		return common.NewAPIError(http.StatusBadRequest, common.ErrActivationTokenInvalid)
	}
	user, err := s.userRepo.FindByID(tok.UserID)
	if err != nil {
		return common.NewAPIError(http.StatusNotFound, common.ErrAuthUserNotFound)
	}
	if err := s.userRepo.UpdateFields(user.ID, map[string]any{
		"status": model.UserStatusActive,
	}); err != nil {
		return common.NewAPIError(http.StatusInternalServerError, common.ErrDatabaseError)
	}
	_ = s.userRepo.ConsumeVerificationToken(tok.ID)

	if s.bus != nil {
		s.bus.Emit(model.EventLevelNormal, model.EventCategoryUser, "event.auth.user_activated", map[string]any{
			"username": user.Username,
		}, nil, "user", user.Username)
	}
	return nil
}

func (s *AuthService) RequestPasswordReset(email, powPayload string) error {
	if s.getStringConfig("smtp_enabled", "false") != "true" {
		return common.NewAPIError(http.StatusBadRequest, common.ErrSMTPNotConfigured)
	}
	if strings.TrimSpace(powPayload) == "" {
		return common.NewAPIError(http.StatusBadRequest, common.ErrAuthCaptchaRequired)
	}
	if !s.captcha.Verify(powPayload) {
		return common.NewAPIError(http.StatusBadRequest, common.ErrAuthCaptchaFailed)
	}

	email = strings.TrimSpace(email)
	if _, err := mail.ParseAddress(email); err != nil {
		return common.NewAPIError(http.StatusBadRequest, common.ErrAuthEmailInvalid)
	}

	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return common.NewAPIError(http.StatusInternalServerError, common.ErrDatabaseError)
	}
	if user.Status == model.UserStatusBanned || user.Status == model.UserStatusDisabled {
		return nil
	}

	rawToken, err := security.GenerateRandomToken(32)
	if err != nil {
		return common.NewAPIError(http.StatusInternalServerError, common.ErrInternalServerError)
	}
	now := time.Now().UTC()
	vToken := &model.VerificationToken{
		UserID:    user.ID,
		TokenHash: security.HashToken(rawToken),
		Purpose:   "password_reset",
		ExpiresAt: now.Add(15 * time.Minute),
		CreatedAt: now,
	}
	if err := s.userRepo.CreateVerificationToken(vToken); err != nil {
		return common.NewAPIError(http.StatusInternalServerError, common.ErrDatabaseError)
	}

	baseURL := strings.TrimRight(s.getStringConfig("base_url", s.baseCfg.Site.BaseURL), "/")
	siteName := s.getStringConfig("server_name", s.baseCfg.Site.Name)
	link := baseURL + "/dashboard?mode=reset&token=" + rawToken
	subj, htmlBody, textBody := mailer.RenderPasswordResetEmail(siteName, user.Username, link)
	if err := s.buildMailer().SendHTML(email, subj, htmlBody, textBody); err != nil {
		return common.NewAPIError(http.StatusBadGateway, common.ErrSMTPTestFailed)
	}

	if s.bus != nil {
		s.bus.Emit(model.EventLevelNormal, model.EventCategorySecurity, "event.auth.password_reset_requested", map[string]any{
			"username": user.Username,
		}, nil, "user", user.Username)
	}
	return nil
}

func (s *AuthService) ResetPassword(rawToken, newPassword string) error {
	rawToken = strings.TrimSpace(rawToken)
	if rawToken == "" {
		return common.NewAPIError(http.StatusBadRequest, common.ErrResetTokenInvalid)
	}
	if len(newPassword) < 8 {
		return common.NewAPIError(http.StatusBadRequest, common.ErrAuthPasswordTooShort)
	}
	tok, err := s.userRepo.FindValidVerificationToken("password_reset", security.HashToken(rawToken))
	if err != nil {
		return common.NewAPIError(http.StatusBadRequest, common.ErrResetTokenInvalid)
	}
	user, err := s.userRepo.FindByID(tok.UserID)
	if err != nil {
		return common.NewAPIError(http.StatusNotFound, common.ErrAuthUserNotFound)
	}

	hash, err := security.HashPassword(newPassword)
	if err != nil {
		return common.NewAPIError(http.StatusInternalServerError, common.ErrInternalServerError)
	}
	if err := s.userRepo.UpdateFields(user.ID, map[string]any{
		"password_hash":         hash,
		"failed_login_attempts": 0,
		"lockout_until":         nil,
	}); err != nil {
		return common.NewAPIError(http.StatusInternalServerError, common.ErrDatabaseError)
	}
	_ = s.userRepo.IncrementTokenVersionAndRevokeSessions(user.ID)
	_ = s.userRepo.ConsumeVerificationToken(tok.ID)

	if s.bus != nil {
		s.bus.Emit(model.EventLevelNormal, model.EventCategorySecurity, "event.auth.password_reset", map[string]any{
			"username": user.Username,
		}, nil, "user", user.Username)
	}
	return nil
}

func (s *AuthService) UpdateProfile(user *model.User, nickname, email string) (UserDTO, error) {
	updates := map[string]any{
		"nickname": strings.TrimSpace(nickname),
	}
	email = strings.TrimSpace(email)
	if email != "" {
		if _, err := mail.ParseAddress(email); err != nil {
			return UserDTO{}, common.NewAPIError(http.StatusBadRequest, common.ErrAuthEmailInvalid)
		}
		if existing, err := s.userRepo.FindByEmail(email); err == nil && existing.ID != user.ID {
			return UserDTO{}, common.NewAPIError(http.StatusBadRequest, common.ErrAuthEmailTaken)
		}
		updates["email"] = &email
	} else {
		updates["email"] = nil
	}

	if err := s.userRepo.UpdateFields(user.ID, updates); err != nil {
		return UserDTO{}, common.SanitizeDBError(err)
	}
	fresh, err := s.userRepo.FindByID(user.ID)
	if err != nil {
		return UserDTO{}, common.NewAPIError(http.StatusInternalServerError, common.ErrDatabaseError)
	}
	return ToUserDTO(fresh), nil
}

func (s *AuthService) ChangePassword(user *model.User, oldPassword, newPassword string) error {
	if len(newPassword) < 8 {
		return common.NewAPIError(http.StatusBadRequest, common.ErrAuthPasswordTooShort)
	}
	ok, err := security.VerifyPassword(oldPassword, user.PasswordHash)
	if err != nil || !ok {
		return common.NewAPIError(http.StatusBadRequest, common.ErrAuthOldPasswordWrong)
	}
	hash, err := security.HashPassword(newPassword)
	if err != nil {
		return common.NewAPIError(http.StatusInternalServerError, common.ErrInternalServerError)
	}
	if err := s.userRepo.UpdateFields(user.ID, map[string]any{
		"password_hash": hash,
	}); err != nil {
		return common.NewAPIError(http.StatusInternalServerError, common.ErrDatabaseError)
	}
	_ = s.userRepo.IncrementTokenVersionAndRevokeSessions(user.ID)
	return nil
}

func (s *AuthService) RevokeAllSessions(user *model.User) error {
	return s.userRepo.IncrementTokenVersionAndRevokeSessions(user.ID)
}

func (s *AuthService) getStringConfig(key, fallback string) string {
	val, err := s.cfgRepo.Get(key)
	if err != nil || strings.TrimSpace(val) == "" {
		return fallback
	}
	return val
}

func (s *AuthService) getIntConfig(key string, fallback int) int {
	val, err := s.cfgRepo.Get(key)
	if err != nil || strings.TrimSpace(val) == "" {
		return fallback
	}
	n, err := strconv.Atoi(strings.TrimSpace(val))
	if err != nil {
		return fallback
	}
	return n
}

func (s *AuthService) getTokenTTLs() (accessTTL, refreshTTL time.Duration) {
	accessMin := s.getIntConfig("jwt_access_minutes", s.baseCfg.Security.JWTAccessMinutes)
	if accessMin <= 0 {
		accessMin = 120
	}
	refreshDays := s.getIntConfig("jwt_refresh_days", s.baseCfg.Security.JWTRefreshDays)
	if refreshDays <= 0 {
		refreshDays = 30
	}
	return time.Duration(accessMin) * time.Minute, time.Duration(refreshDays) * 24 * time.Hour
}

func (s *AuthService) buildMailer() *mailer.Mailer {
	m, _ := s.cfgRepo.GetMap()
	port, _ := strconv.Atoi(m["smtp_port"])
	if port <= 0 {
		port = 465
	}
	return mailer.NewMailer(mailer.Config{
		Enabled:    m["smtp_enabled"] == "true",
		Host:       m["smtp_host"],
		Port:       port,
		Encryption: m["smtp_encryption"],
		From:       m["smtp_from"],
		FromName:   m["smtp_from_name"],
		User:       m["smtp_user"],
		Password:   m["smtp_password"],
	})
}
