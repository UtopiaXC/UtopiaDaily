package model

import "time"

type Permission struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Code      string    `gorm:"size:128;uniqueIndex;not null" json:"code"`
	GroupCode string    `gorm:"size:64;not null;index" json:"group_code"`
	CreatedAt time.Time `json:"created_at"`
}

func (Permission) TableName() string {
	return "permissions"
}

type Role struct {
	ID          uint         `gorm:"primaryKey" json:"id"`
	Name        string       `gorm:"size:64;uniqueIndex;not null" json:"name"`
	Description string       `gorm:"size:255" json:"description"`
	IsSystem    bool         `gorm:"default:false;not null" json:"is_system"`
	Permissions []Permission `gorm:"many2many:role_permissions;constraint:OnDelete:CASCADE;" json:"permissions,omitempty"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

func (Role) TableName() string {
	return "roles"
}

type RolePermission struct {
	RoleID       uint `gorm:"primaryKey"`
	PermissionID uint `gorm:"primaryKey"`
}

func (RolePermission) TableName() string {
	return "role_permissions"
}

const (
	UserStatusActive   = "active"
	UserStatusInactive = "inactive"
	UserStatusFrozen   = "frozen"
	UserStatusBanned   = "banned"
	UserStatusDisabled = "disabled"
)

type User struct {
	ID                  uint       `gorm:"primaryKey" json:"id"`
	Username            string     `gorm:"size:64;uniqueIndex;not null" json:"username"`
	Nickname            string     `gorm:"size:64" json:"nickname"`
	Email               *string    `gorm:"size:128;uniqueIndex" json:"email"`
	PasswordHash        string     `gorm:"size:255;not null" json:"-"`
	RoleID              uint       `gorm:"not null;index" json:"role_id"`
	Role                *Role      `gorm:"foreignKey:RoleID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"role,omitempty"`
	Status              string     `gorm:"size:20;default:active;not null;index" json:"status"`
	TokenVersion        int64      `gorm:"default:1;not null" json:"-"`
	FailedLoginAttempts int        `gorm:"default:0;not null" json:"failed_login_attempts"`
	LockoutUntil        *time.Time `json:"lockout_until,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

func (User) TableName() string {
	return "users"
}

type UserSession struct {
	ID           string    `gorm:"primaryKey;size:64"`
	UserID       uint      `gorm:"not null;index"`
	User         *User     `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE;"`
	TokenVersion int64     `gorm:"not null"`
	IP           string    `gorm:"size:64"`
	UserAgent    string    `gorm:"size:255"`
	RememberMe   bool      `gorm:"default:false;not null"`
	ExpiresAt    time.Time `gorm:"index;not null"`
	LastSeenAt   time.Time `gorm:"not null"`
	CreatedAt    time.Time `gorm:"not null"`
}

func (UserSession) TableName() string {
	return "user_sessions"
}

type VerificationToken struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"not null;index"`
	User      *User     `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE;"`
	TokenHash string    `gorm:"size:64;uniqueIndex;not null"`
	Purpose   string    `gorm:"size:32;not null;index"`
	ExpiresAt time.Time `gorm:"not null;index"`
	UsedAt    *time.Time
	CreatedAt time.Time `gorm:"not null"`
}

func (VerificationToken) TableName() string {
	return "verification_tokens"
}
