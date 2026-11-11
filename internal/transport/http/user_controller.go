package http

import (
	"net/http"
	"strconv"

	"github.com/UtopiaXC/UtopiaDaily/internal/common"
	"github.com/UtopiaXC/UtopiaDaily/internal/security"
	"github.com/UtopiaXC/UtopiaDaily/internal/service"
	"github.com/gin-gonic/gin"
)

type UserController struct {
	userSvc *service.UserService
}

func NewUserController(userSvc *service.UserService) *UserController {
	return &UserController{userSvc: userSvc}
}

func (ctl *UserController) ListUsers(c *gin.Context) {
	list, err := ctl.userSvc.ListUsers()
	if err != nil {
		common.RespondFromError(c, err)
		return
	}
	common.RespondSuccess(c, list)
}

func (ctl *UserController) CreateUser(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Nickname string `json:"nickname"`
		Email    string `json:"email"`
		Password string `json:"password"`
		RoleID   uint   `json:"role_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.RespondError(c, http.StatusBadRequest, common.ErrValidationFailed)
		return
	}
	op, _ := security.CurrentUser(c)
	opName := "system"
	if op != nil {
		opName = op.Username
	}
	dto, err := ctl.userSvc.CreateUser(opName, req.Username, req.Nickname, req.Email, req.Password, req.RoleID)
	if err != nil {
		common.RespondFromError(c, err)
		return
	}
	common.RespondSuccess(c, dto)
}

func (ctl *UserController) UpdateUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		common.RespondError(c, http.StatusBadRequest, common.ErrValidationFailed)
		return
	}
	var req struct {
		Nickname string `json:"nickname"`
		Email    string `json:"email"`
		RoleID   uint   `json:"role_id"`
		Status   string `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.RespondError(c, http.StatusBadRequest, common.ErrValidationFailed)
		return
	}
	op, _ := security.CurrentUser(c)
	var opID uint
	opName := "system"
	if op != nil {
		opID = op.ID
		opName = op.Username
	}
	dto, err := ctl.userSvc.UpdateUser(opName, opID, uint(id), req.Nickname, req.Email, req.RoleID, req.Status)
	if err != nil {
		common.RespondFromError(c, err)
		return
	}
	common.RespondSuccess(c, dto)
}

func (ctl *UserController) ResetUserPassword(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		common.RespondError(c, http.StatusBadRequest, common.ErrValidationFailed)
		return
	}
	var req struct {
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.RespondError(c, http.StatusBadRequest, common.ErrValidationFailed)
		return
	}
	op, _ := security.CurrentUser(c)
	opName := "system"
	if op != nil {
		opName = op.Username
	}
	if err := ctl.userSvc.ResetUserPassword(opName, uint(id), req.Password); err != nil {
		common.RespondFromError(c, err)
		return
	}
	common.RespondSuccess(c, gin.H{"status": "ok"})
}

func (ctl *UserController) BanUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		common.RespondError(c, http.StatusBadRequest, common.ErrValidationFailed)
		return
	}
	op, _ := security.CurrentUser(c)
	var opID uint
	opName := "system"
	if op != nil {
		opID = op.ID
		opName = op.Username
	}
	if err := ctl.userSvc.BanUser(opName, opID, uint(id)); err != nil {
		common.RespondFromError(c, err)
		return
	}
	common.RespondSuccess(c, gin.H{"status": "ok"})
}

func (ctl *UserController) UnfreezeUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		common.RespondError(c, http.StatusBadRequest, common.ErrValidationFailed)
		return
	}
	op, _ := security.CurrentUser(c)
	opName := "system"
	if op != nil {
		opName = op.Username
	}
	if err := ctl.userSvc.UnfreezeUser(opName, uint(id)); err != nil {
		common.RespondFromError(c, err)
		return
	}
	common.RespondSuccess(c, gin.H{"status": "ok"})
}

func (ctl *UserController) DeleteUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		common.RespondError(c, http.StatusBadRequest, common.ErrValidationFailed)
		return
	}
	op, _ := security.CurrentUser(c)
	var opID uint
	opName := "system"
	if op != nil {
		opID = op.ID
		opName = op.Username
	}
	if err := ctl.userSvc.DeleteUser(opName, opID, uint(id)); err != nil {
		common.RespondFromError(c, err)
		return
	}
	common.RespondSuccess(c, gin.H{"status": "ok"})
}

func (ctl *UserController) ListRoles(c *gin.Context) {
	roles, err := ctl.userSvc.ListRoles()
	if err != nil {
		common.RespondError(c, http.StatusInternalServerError, common.ErrDatabaseError)
		return
	}
	common.RespondSuccess(c, roles)
}

func (ctl *UserController) ListPermissions(c *gin.Context) {
	perms, err := ctl.userSvc.ListPermissions()
	if err != nil {
		common.RespondError(c, http.StatusInternalServerError, common.ErrDatabaseError)
		return
	}
	common.RespondSuccess(c, perms)
}

func (ctl *UserController) CreateRole(c *gin.Context) {
	var req struct {
		Name        string   `json:"name"`
		Description string   `json:"description"`
		Permissions []string `json:"permissions"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.RespondError(c, http.StatusBadRequest, common.ErrValidationFailed)
		return
	}
	role, err := ctl.userSvc.CreateRole(req.Name, req.Description, req.Permissions)
	if err != nil {
		common.RespondFromError(c, err)
		return
	}
	common.RespondSuccess(c, role)
}

func (ctl *UserController) UpdateRole(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		common.RespondError(c, http.StatusBadRequest, common.ErrValidationFailed)
		return
	}
	var req struct {
		Name        string   `json:"name"`
		Description string   `json:"description"`
		Permissions []string `json:"permissions"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.RespondError(c, http.StatusBadRequest, common.ErrValidationFailed)
		return
	}
	if err := ctl.userSvc.UpdateRole(uint(id), req.Name, req.Description, req.Permissions); err != nil {
		common.RespondFromError(c, err)
		return
	}
	common.RespondSuccess(c, gin.H{"status": "ok"})
}

func (ctl *UserController) DeleteRole(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		common.RespondError(c, http.StatusBadRequest, common.ErrValidationFailed)
		return
	}
	if err := ctl.userSvc.DeleteRole(uint(id)); err != nil {
		common.RespondFromError(c, err)
		return
	}
	common.RespondSuccess(c, gin.H{"status": "ok"})
}
