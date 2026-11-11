package repository

import (
	"strings"
	"time"

	"github.com/UtopiaXC/UtopiaDaily/internal/model"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) CountUsers() (int64, error) {
	var count int64
	err := r.db.Model(&model.User{}).Count(&count).Error
	return count, err
}

func (r *UserRepository) HasAdmin() (bool, error) {
	var count int64
	err := r.db.Model(&model.User{}).
		Joins("JOIN roles ON roles.id = users.role_id").
		Where("roles.name = ?", "admin").
		Count(&count).Error
	return count > 0, err
}

func (r *UserRepository) CountActiveAdmins(excludeID uint) (int64, error) {
	var count int64
	q := r.db.Model(&model.User{}).
		Joins("JOIN roles ON roles.id = users.role_id").
		Where("roles.name = ? AND users.status = ?", "admin", model.UserStatusActive)
	if excludeID > 0 {
		q = q.Where("users.id <> ?", excludeID)
	}
	err := q.Count(&count).Error
	return count, err
}

func (r *UserRepository) FindByID(id uint) (*model.User, error) {
	var u model.User
	err := r.db.Preload("Role.Permissions").Where("id = ?", id).First(&u).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) FindByAccount(account string) (*model.User, error) {
	q := strings.TrimSpace(account)
	var u model.User
	err := r.db.Preload("Role.Permissions").
		Where("LOWER(username) = LOWER(?) OR LOWER(email) = LOWER(?)", q, q).
		First(&u).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) FindByUsername(username string) (*model.User, error) {
	var u model.User
	err := r.db.Preload("Role.Permissions").
		Where("LOWER(username) = LOWER(?)", strings.TrimSpace(username)).
		First(&u).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) FindByEmail(email string) (*model.User, error) {
	var u model.User
	err := r.db.Preload("Role.Permissions").
		Where("LOWER(email) = LOWER(?)", strings.TrimSpace(email)).
		First(&u).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) Create(u *model.User) error {
	return r.db.Create(u).Error
}

func (r *UserRepository) UpdateFields(id uint, updates map[string]any) error {
	updates["updated_at"] = time.Now().UTC()
	return r.db.Model(&model.User{}).Where("id = ?", id).Updates(updates).Error
}

func (r *UserRepository) IncrementTokenVersionAndRevokeSessions(userID uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.User{}).Where("id = ?", userID).Updates(map[string]any{
			"token_version": gorm.Expr("token_version + 1"),
			"updated_at":    time.Now().UTC(),
		}).Error; err != nil {
			return err
		}
		return tx.Where("user_id = ?", userID).Delete(&model.UserSession{}).Error
	})
}

func (r *UserRepository) Delete(id uint) error {
	return r.db.Delete(&model.User{}, id).Error
}

func (r *UserRepository) ListUsers() ([]model.User, error) {
	var users []model.User
	err := r.db.Preload("Role").Order("id ASC").Find(&users).Error
	return users, err
}

func (r *UserRepository) CreateVerificationToken(t *model.VerificationToken) error {
	return r.db.Create(t).Error
}

func (r *UserRepository) FindValidVerificationToken(purpose, tokenHash string) (*model.VerificationToken, error) {
	var tok model.VerificationToken
	err := r.db.Where("purpose = ? AND token_hash = ? AND used_at IS NULL AND expires_at > ?",
		purpose, tokenHash, time.Now().UTC()).First(&tok).Error
	if err != nil {
		return nil, err
	}
	return &tok, nil
}

func (r *UserRepository) ConsumeVerificationToken(id uint) error {
	now := time.Now().UTC()
	return r.db.Model(&model.VerificationToken{}).Where("id = ?", id).Update("used_at", now).Error
}

type SessionRepository struct {
	db *gorm.DB
}

func NewSessionRepository(db *gorm.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

func (s *SessionRepository) CreateSession(sess *model.UserSession) error {
	return s.db.Create(sess).Error
}

func (s *SessionRepository) GetSessionWithUser(sessionID string) (*model.UserSession, error) {
	var sess model.UserSession
	err := s.db.Preload("User.Role.Permissions").Where("id = ?", sessionID).First(&sess).Error
	if err != nil {
		return nil, err
	}
	return &sess, nil
}

func (s *SessionRepository) UpdateSessionLastSeen(sessionID string, seenAt time.Time) error {
	return s.db.Model(&model.UserSession{}).Where("id = ?", sessionID).Update("last_seen_at", seenAt).Error
}

func (s *SessionRepository) RotateSession(oldSessionID string, newSess *model.UserSession) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("id = ?", oldSessionID).Delete(&model.UserSession{}).Error; err != nil {
			return err
		}
		return tx.Create(newSess).Error
	})
}

func (s *SessionRepository) DeleteSession(sessionID string) error {
	return s.db.Where("id = ?", sessionID).Delete(&model.UserSession{}).Error
}

func (s *SessionRepository) DeleteUserSessions(userID uint) error {
	return s.db.Where("user_id = ?", userID).Delete(&model.UserSession{}).Error
}
