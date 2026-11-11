package http

import (
	"net/http"
	"strings"

	"github.com/UtopiaXC/UtopiaDaily/internal/common"
	"github.com/UtopiaXC/UtopiaDaily/internal/security"
	"github.com/UtopiaXC/UtopiaDaily/internal/service"
	"github.com/gin-gonic/gin"
)

type AuthController struct {
	authSvc *service.AuthService
}

func NewAuthController(authSvc *service.AuthService) *AuthController {
	return &AuthController{authSvc: authSvc}
}

func (ctl *AuthController) SetupStatus(c *gin.Context) {
	required, err := ctl.authSvc.IsSetupRequired()
	if err != nil {
		common.RespondError(c, http.StatusInternalServerError, common.ErrDatabaseError)
		return
	}
	common.RespondSuccess(c, gin.H{"setup_required": required})
}

func (ctl *AuthController) SetupAdmin(c *gin.Context) {
	var req struct {
		SiteName        string `json:"site_name"`
		SiteDescription string `json:"site_description"`
		BaseURL         string `json:"base_url"`
		DefaultLocale   string `json:"default_locale"`
		SMTPEnabled     bool   `json:"smtp_enabled"`
		SMTPHost        string `json:"smtp_host"`
		SMTPPort        int    `json:"smtp_port"`
		SMTPEncryption  string `json:"smtp_encryption"`
		SMTPUser        string `json:"smtp_user"`
		SMTPPassword    string `json:"smtp_password"`
		SMTPFrom        string `json:"smtp_from"`
		SMTPFromName    string `json:"smtp_from_name"`
		AdminUsername   string `json:"admin_username"`
		AdminNickname   string `json:"admin_nickname"`
		AdminEmail      string `json:"admin_email"`
		AdminPassword   string `json:"admin_password"`
		Username        string `json:"username"`
		Nickname        string `json:"nickname"`
		Email           string `json:"email"`
		Password        string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.RespondError(c, http.StatusBadRequest, common.ErrValidationFailed)
		return
	}
	username := firstNonEmpty(req.AdminUsername, req.Username)
	nickname := firstNonEmpty(req.AdminNickname, req.Nickname)
	email := firstNonEmpty(req.AdminEmail, req.Email)
	password := firstNonEmpty(req.AdminPassword, req.Password)

	resp, err := ctl.authSvc.SetupAdmin(service.SetupInput{
		SiteName:        req.SiteName,
		SiteDescription: req.SiteDescription,
		BaseURL:         req.BaseURL,
		DefaultLocale:   req.DefaultLocale,
		SMTPEnabled:     req.SMTPEnabled,
		SMTPHost:        req.SMTPHost,
		SMTPPort:        req.SMTPPort,
		SMTPEncryption:  req.SMTPEncryption,
		SMTPUser:        req.SMTPUser,
		SMTPPassword:    req.SMTPPassword,
		SMTPFrom:        req.SMTPFrom,
		SMTPFromName:    req.SMTPFromName,
		Username:        username,
		Nickname:        nickname,
		Email:           email,
		Password:        password,
	}, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		common.RespondFromError(c, err)
		return
	}
	common.RespondSuccess(c, resp)
}

func (ctl *AuthController) LoginPolicy(c *gin.Context) {
	common.RespondSuccess(c, ctl.authSvc.GetLoginPolicy())
}

func (ctl *AuthController) Captcha(c *gin.Context) {
	ch, err := ctl.authSvc.GenerateCaptcha()
	if err != nil {
		common.RespondError(c, http.StatusInternalServerError, common.ErrInternalServerError)
		return
	}
	common.RespondSuccess(c, ch)
}

func (ctl *AuthController) Login(c *gin.Context) {
	var req struct {
		Account    string `json:"account"`
		Username   string `json:"username"`
		Password   string `json:"password"`
		PowPayload string `json:"pow_payload"`
		RememberMe bool   `json:"remember_me"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.RespondError(c, http.StatusBadRequest, common.ErrValidationFailed)
		return
	}
	account := strings.TrimSpace(req.Account)
	if account == "" {
		account = strings.TrimSpace(req.Username)
	}
	resp, err := ctl.authSvc.Login(account, req.Password, req.PowPayload, req.RememberMe, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		common.RespondFromError(c, err)
		return
	}
	common.RespondSuccess(c, resp)
}

func (ctl *AuthController) Refresh(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.RespondError(c, http.StatusUnauthorized, common.ErrAuthUnauthorized)
		return
	}
	resp, err := ctl.authSvc.RefreshToken(req.RefreshToken, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		common.RespondFromError(c, err)
		return
	}
	common.RespondSuccess(c, resp)
}

func (ctl *AuthController) Logout(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	_ = c.ShouldBindJSON(&req)
	sid := c.GetString(security.CtxSessionIDKey)
	_ = ctl.authSvc.Logout(sid, req.RefreshToken)
	common.RespondSuccess(c, gin.H{"status": "ok"})
}

func (ctl *AuthController) Register(c *gin.Context) {
	var req struct {
		Username   string `json:"username"`
		Nickname   string `json:"nickname"`
		Email      string `json:"email"`
		Password   string `json:"password"`
		PowPayload string `json:"pow_payload"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.RespondError(c, http.StatusBadRequest, common.ErrValidationFailed)
		return
	}
	if err := ctl.authSvc.Register(req.Username, req.Nickname, req.Email, req.Password, req.PowPayload); err != nil {
		common.RespondFromError(c, err)
		return
	}
	common.RespondSuccess(c, gin.H{"status": "ok"})
}

func (ctl *AuthController) Activate(c *gin.Context) {
	var req struct {
		Token string `json:"token"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.RespondError(c, http.StatusBadRequest, common.ErrValidationFailed)
		return
	}
	if err := ctl.authSvc.ActivateAccount(req.Token); err != nil {
		common.RespondFromError(c, err)
		return
	}
	common.RespondSuccess(c, gin.H{"status": "ok"})
}

func (ctl *AuthController) ForgotPassword(c *gin.Context) {
	var req struct {
		Email      string `json:"email"`
		PowPayload string `json:"pow_payload"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.RespondError(c, http.StatusBadRequest, common.ErrValidationFailed)
		return
	}
	if err := ctl.authSvc.RequestPasswordReset(req.Email, req.PowPayload); err != nil {
		common.RespondFromError(c, err)
		return
	}
	common.RespondSuccess(c, gin.H{"status": "ok"})
}

func (ctl *AuthController) ResetPassword(c *gin.Context) {
	var req struct {
		Token       string `json:"token"`
		NewPassword string `json:"new_password"`
		Password    string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.RespondError(c, http.StatusBadRequest, common.ErrValidationFailed)
		return
	}
	newPass := firstNonEmpty(req.NewPassword, req.Password)
	if err := ctl.authSvc.ResetPassword(req.Token, newPass); err != nil {
		common.RespondFromError(c, err)
		return
	}
	common.RespondSuccess(c, gin.H{"status": "ok"})
}

func (ctl *AuthController) Me(c *gin.Context) {
	user, ok := security.CurrentUser(c)
	if !ok {
		common.RespondError(c, http.StatusUnauthorized, common.ErrAuthUnauthorized)
		return
	}
	common.RespondSuccess(c, service.ToUserDTO(user))
}

func (ctl *AuthController) UpdateProfile(c *gin.Context) {
	user, ok := security.CurrentUser(c)
	if !ok {
		common.RespondError(c, http.StatusUnauthorized, common.ErrAuthUnauthorized)
		return
	}
	var req struct {
		Nickname string `json:"nickname"`
		Email    string `json:"email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.RespondError(c, http.StatusBadRequest, common.ErrValidationFailed)
		return
	}
	dto, err := ctl.authSvc.UpdateProfile(user, req.Nickname, req.Email)
	if err != nil {
		common.RespondFromError(c, err)
		return
	}
	common.RespondSuccess(c, dto)
}

func (ctl *AuthController) ChangePassword(c *gin.Context) {
	user, ok := security.CurrentUser(c)
	if !ok {
		common.RespondError(c, http.StatusUnauthorized, common.ErrAuthUnauthorized)
		return
	}
	var req struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.RespondError(c, http.StatusBadRequest, common.ErrValidationFailed)
		return
	}
	if err := ctl.authSvc.ChangePassword(user, req.OldPassword, req.NewPassword); err != nil {
		common.RespondFromError(c, err)
		return
	}
	common.RespondSuccess(c, gin.H{"status": "ok"})
}

func (ctl *AuthController) LogoutAll(c *gin.Context) {
	user, ok := security.CurrentUser(c)
	if !ok {
		common.RespondError(c, http.StatusUnauthorized, common.ErrAuthUnauthorized)
		return
	}
	if err := ctl.authSvc.RevokeAllSessions(user); err != nil {
		common.RespondError(c, http.StatusInternalServerError, common.ErrDatabaseError)
		return
	}
	common.RespondSuccess(c, gin.H{"status": "ok"})
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
