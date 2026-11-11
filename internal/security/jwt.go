package security

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strconv"
	"time"

	"github.com/UtopiaXC/UtopiaDaily/internal/model"
	"github.com/golang-jwt/jwt/v5"
)

type SessionStore interface {
	CreateSession(sess *model.UserSession) error
	GetSessionWithUser(sessionID string) (*model.UserSession, error)
	UpdateSessionLastSeen(sessionID string, seenAt time.Time) error
	RotateSession(oldSessionID string, newSess *model.UserSession) error
	DeleteSession(sessionID string) error
	DeleteUserSessions(userID uint) error
}

type TokenPair struct {
	AccessToken  string `json:"token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

type Claims struct {
	UserID       uint   `json:"uid"`
	SessionID    string `json:"sid"`
	TokenVersion int64  `json:"ver"`
	jwt.RegisteredClaims
}

type JWTManager struct {
	secret []byte
	store  SessionStore
}

func NewJWTManager(secret string, store SessionStore) *JWTManager {
	return &JWTManager{
		secret: []byte(secret),
		store:  store,
	}
}

func GenerateRandomToken(numBytes int) (string, error) {
	b := make([]byte, numBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func HashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func (jm *JWTManager) IssueTokens(user *model.User, ip, userAgent string, rememberMe bool, accessTTL, refreshTTL time.Duration) (*TokenPair, error) {
	rawRefresh, err := GenerateRandomToken(32)
	if err != nil {
		return nil, err
	}
	sessionID := HashToken(rawRefresh)
	now := time.Now()

	sess := &model.UserSession{
		ID:           sessionID,
		UserID:       user.ID,
		TokenVersion: user.TokenVersion,
		IP:           ip,
		UserAgent:    userAgent,
		RememberMe:   rememberMe,
		ExpiresAt:    now.Add(refreshTTL),
		LastSeenAt:   now,
		CreatedAt:    now,
	}
	if err := jm.store.CreateSession(sess); err != nil {
		return nil, err
	}

	accessToken, err := jm.signAccessToken(user, sessionID, now, accessTTL)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: rawRefresh,
		ExpiresIn:    int64(accessTTL.Seconds()),
	}, nil
}

func (jm *JWTManager) VerifyAccessToken(tokenStr string) (*model.User, string, error) {
	claims := &Claims{}
	parsed, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jm.secret, nil
	})
	if err != nil || !parsed.Valid {
		return nil, "", errors.New("invalid or expired token")
	}

	sess, err := jm.store.GetSessionWithUser(claims.SessionID)
	if err != nil || sess == nil || sess.User == nil {
		return nil, "", errors.New("session not found")
	}
	if time.Now().After(sess.ExpiresAt) {
		_ = jm.store.DeleteSession(claims.SessionID)
		return nil, "", errors.New("session expired")
	}
	if sess.User.Status != model.UserStatusActive {
		return nil, "", errors.New("user not active")
	}
	if sess.User.TokenVersion != claims.TokenVersion || sess.User.TokenVersion != sess.TokenVersion {
		_ = jm.store.DeleteSession(claims.SessionID)
		return nil, "", errors.New("token revoked")
	}

	if time.Since(sess.LastSeenAt) > 10*time.Minute {
		_ = jm.store.UpdateSessionLastSeen(claims.SessionID, time.Now())
	}

	return sess.User, claims.SessionID, nil
}

func (jm *JWTManager) RefreshTokens(rawRefresh string, ip, userAgent string, accessTTL, refreshTTL time.Duration) (*TokenPair, *model.User, error) {
	oldSessionID := HashToken(rawRefresh)
	sess, err := jm.store.GetSessionWithUser(oldSessionID)
	if err != nil || sess == nil || sess.User == nil {
		return nil, nil, errors.New("invalid refresh token")
	}
	now := time.Now()
	if now.After(sess.ExpiresAt) {
		_ = jm.store.DeleteSession(oldSessionID)
		return nil, nil, errors.New("refresh token expired")
	}
	if sess.User.Status != model.UserStatusActive {
		return nil, nil, errors.New("user not active")
	}
	if sess.User.TokenVersion != sess.TokenVersion {
		_ = jm.store.DeleteSession(oldSessionID)
		return nil, nil, errors.New("token revoked")
	}

	newRawRefresh, err := GenerateRandomToken(32)
	if err != nil {
		return nil, nil, err
	}
	newSessionID := HashToken(newRawRefresh)

	newRefreshTTL := refreshTTL
	if !sess.RememberMe {
		newRefreshTTL = 24 * time.Hour
	}

	newSess := &model.UserSession{
		ID:           newSessionID,
		UserID:       sess.User.ID,
		TokenVersion: sess.User.TokenVersion,
		IP:           ip,
		UserAgent:    userAgent,
		RememberMe:   sess.RememberMe,
		ExpiresAt:    now.Add(newRefreshTTL),
		LastSeenAt:   now,
		CreatedAt:    now,
	}
	if err := jm.store.RotateSession(oldSessionID, newSess); err != nil {
		return nil, nil, err
	}

	accessToken, err := jm.signAccessToken(sess.User, newSessionID, now, accessTTL)
	if err != nil {
		return nil, nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: newRawRefresh,
		ExpiresIn:    int64(accessTTL.Seconds()),
	}, sess.User, nil
}

func (jm *JWTManager) RevokeSession(sessionID string) error {
	if sessionID == "" {
		return nil
	}
	return jm.store.DeleteSession(sessionID)
}

func (jm *JWTManager) RevokeByRefreshToken(rawRefresh string) error {
	if rawRefresh == "" {
		return nil
	}
	return jm.store.DeleteSession(HashToken(rawRefresh))
}

func (jm *JWTManager) RevokeAllUserSessions(userID uint) error {
	return jm.store.DeleteUserSessions(userID)
}

func (jm *JWTManager) signAccessToken(user *model.User, sessionID string, now time.Time, ttl time.Duration) (string, error) {
	jti, _ := GenerateRandomToken(12)
	claims := Claims{
		UserID:       user.ID,
		SessionID:    sessionID,
		TokenVersion: user.TokenVersion,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatUint(uint64(user.ID), 10),
			Issuer:    "utopia-daily",
			ID:        jti,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jm.secret)
}
