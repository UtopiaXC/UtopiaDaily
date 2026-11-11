package security

import (
	"net/http"
	"strings"

	"github.com/UtopiaXC/UtopiaDaily/internal/common"
	"github.com/UtopiaXC/UtopiaDaily/internal/model"
	"github.com/gin-gonic/gin"
)

const (
	CtxUserKey        = "current_user"
	CtxUserIDKey      = "current_user_id"
	CtxSessionIDKey   = "current_session_id"
	CtxPermissionsKey = "current_permissions"
)

func AuthMiddleware(jwtMgr *JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			tokenStr := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
			if tokenStr != "" {
				if user, sid, err := jwtMgr.VerifyAccessToken(tokenStr); err == nil && user != nil {
					permSet := make(map[string]bool)
					if user.Role != nil {
						for _, p := range user.Role.Permissions {
							permSet[p.Code] = true
						}
					}
					c.Set(CtxUserKey, user)
					c.Set(CtxUserIDKey, user.ID)
					c.Set(CtxSessionIDKey, sid)
					c.Set(CtxPermissionsKey, permSet)
				}
			}
		}
		c.Next()
	}
}

func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		uVal, exists := c.Get(CtxUserKey)
		if !exists {
			common.RespondError(c, http.StatusUnauthorized, common.ErrAuthUnauthorized)
			c.Abort()
			return
		}
		user, ok := uVal.(*model.User)
		if !ok || user == nil || user.Status != model.UserStatusActive {
			common.RespondError(c, http.StatusUnauthorized, common.ErrAuthUnauthorized)
			c.Abort()
			return
		}
		c.Next()
	}
}

func RequirePermission(code string) gin.HandlerFunc {
	return func(c *gin.Context) {
		uVal, exists := c.Get(CtxUserKey)
		if !exists {
			common.RespondError(c, http.StatusUnauthorized, common.ErrAuthUnauthorized)
			c.Abort()
			return
		}
		user := uVal.(*model.User)
		if user.Role != nil && user.Role.Name == "admin" {
			c.Next()
			return
		}
		pVal, ok := c.Get(CtxPermissionsKey)
		if !ok {
			common.RespondError(c, http.StatusForbidden, common.ErrAuthForbidden)
			c.Abort()
			return
		}
		permSet, _ := pVal.(map[string]bool)
		if !permSet[code] {
			common.RespondError(c, http.StatusForbidden, common.ErrAuthForbidden)
			c.Abort()
			return
		}
		c.Next()
	}
}

func CurrentUser(c *gin.Context) (*model.User, bool) {
	v, ok := c.Get(CtxUserKey)
	if !ok {
		return nil, false
	}
	u, ok := v.(*model.User)
	return u, ok && u != nil
}

func CurrentPermissions(c *gin.Context) map[string]bool {
	v, ok := c.Get(CtxPermissionsKey)
	if !ok {
		return map[string]bool{}
	}
	m, _ := v.(map[string]bool)
	if m == nil {
		return map[string]bool{}
	}
	return m
}

func HasPermission(c *gin.Context, code string) bool {
	u, ok := CurrentUser(c)
	if !ok {
		return false
	}
	if u.Role != nil && u.Role.Name == "admin" {
		return true
	}
	return CurrentPermissions(c)[code]
}
