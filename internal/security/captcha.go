package security

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Challenge struct {
	Algorithm string `json:"algorithm"`
	Challenge string `json:"challenge"`
	Salt      string `json:"salt"`
	Signature string `json:"signature"`
	MaxNumber int64  `json:"maxnumber"`
}

type CaptchaProvider interface {
	Generate(difficulty int64) (*Challenge, error)
	Verify(payloadJSON string) bool
}

type AltchaPayload struct {
	Algorithm string `json:"algorithm"`
	Challenge string `json:"challenge"`
	Number    int64  `json:"number"`
	Salt      string `json:"salt"`
	Signature string `json:"signature"`
}

type AltchaProvider struct {
	secretKey      string
	usedChallenges map[string]time.Time
	mu             sync.Mutex
}

func NewAltchaProvider(secretKey string) *AltchaProvider {
	return &AltchaProvider{
		secretKey:      secretKey,
		usedChallenges: make(map[string]time.Time),
	}
}

func (p *AltchaProvider) Generate(difficulty int64) (*Challenge, error) {
	if difficulty <= 0 {
		difficulty = 50000
	}

	saltBytes := make([]byte, 16)
	if _, err := rand.Read(saltBytes); err != nil {
		return nil, err
	}
	expiresAt := time.Now().Add(5 * time.Minute).Unix()
	salt := fmt.Sprintf("%s?expires=%d", hex.EncodeToString(saltBytes), expiresAt)

	nBig, err := rand.Int(rand.Reader, big.NewInt(difficulty))
	if err != nil {
		return nil, err
	}
	secretNumber := nBig.Int64()

	sum := sha256.Sum256([]byte(fmt.Sprintf("%s%d", salt, secretNumber)))
	challenge := hex.EncodeToString(sum[:])

	mac := hmac.New(sha256.New, []byte(p.secretKey))
	mac.Write([]byte(challenge + ":" + salt))
	signature := hex.EncodeToString(mac.Sum(nil))

	return &Challenge{
		Algorithm: "SHA-256",
		Challenge: challenge,
		Salt:      salt,
		Signature: signature,
		MaxNumber: difficulty,
	}, nil
}

func (p *AltchaProvider) Verify(payloadJSON string) bool {
	raw := strings.TrimSpace(payloadJSON)
	if raw == "" {
		return false
	}
	if !strings.HasPrefix(raw, "{") {
		if decoded, err := base64.StdEncoding.DecodeString(raw); err == nil {
			raw = string(decoded)
		} else if decoded, err := base64.RawStdEncoding.DecodeString(raw); err == nil {
			raw = string(decoded)
		}
	}
	var payload AltchaPayload
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return false
	}
	if payload.Algorithm != "" && !strings.EqualFold(payload.Algorithm, "SHA-256") {
		return false
	}

	mac := hmac.New(sha256.New, []byte(p.secretKey))
	mac.Write([]byte(payload.Challenge + ":" + payload.Salt))
	expectedSig := hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(strings.ToLower(payload.Signature)), []byte(expectedSig)) {
		return false
	}

	if idx := strings.Index(payload.Salt, "?expires="); idx != -1 {
		expStr := payload.Salt[idx+len("?expires="):]
		if exp, err := strconv.ParseInt(expStr, 10, 64); err != nil || time.Now().Unix() > exp {
			return false
		}
	}

	sum := sha256.Sum256([]byte(fmt.Sprintf("%s%d", payload.Salt, payload.Number)))
	if hex.EncodeToString(sum[:]) != strings.ToLower(payload.Challenge) {
		return false
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	now := time.Now()
	if len(p.usedChallenges) > 1000 {
		for k, expTime := range p.usedChallenges {
			if now.After(expTime) {
				delete(p.usedChallenges, k)
			}
		}
	}
	if _, used := p.usedChallenges[payload.Challenge]; used {
		return false
	}
	p.usedChallenges[payload.Challenge] = now.Add(10 * time.Minute)
	return true
}

type LoginAttemptsTracker struct {
	mu          sync.RWMutex
	attempts    map[string]int
	lockouts    map[string]time.Time
	lastAttempt map[string]time.Time
}

func NewLoginAttemptsTracker() *LoginAttemptsTracker {
	t := &LoginAttemptsTracker{
		attempts:    make(map[string]int),
		lockouts:    make(map[string]time.Time),
		lastAttempt: make(map[string]time.Time),
	}
	go t.cleanupLoop()
	return t
}

func (t *LoginAttemptsTracker) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		t.cleanup()
	}
}

func (t *LoginAttemptsTracker) cleanup() {
	t.mu.Lock()
	defer t.mu.Unlock()
	now := time.Now()
	for k, until := range t.lockouts {
		if now.After(until) {
			delete(t.lockouts, k)
			delete(t.attempts, k)
			delete(t.lastAttempt, k)
		}
	}
	for k, last := range t.lastAttempt {
		if now.Sub(last) > time.Hour {
			delete(t.attempts, k)
			delete(t.lastAttempt, k)
			if until, ok := t.lockouts[k]; !ok || now.After(until) {
				delete(t.lockouts, k)
			}
		}
	}
}

func (t *LoginAttemptsTracker) RecordFailure(key string, maxAttempts int, lockout time.Duration) (int, bool) {
	if maxAttempts <= 0 {
		maxAttempts = 5
	}
	if lockout <= 0 {
		lockout = 15 * time.Minute
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	now := time.Now()
	t.attempts[key]++
	t.lastAttempt[key] = now
	count := t.attempts[key]
	if count >= maxAttempts {
		t.lockouts[key] = now.Add(lockout)
		return count, true
	}
	return count, false
}

func (t *LoginAttemptsTracker) IsLocked(key string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	until, ok := t.lockouts[key]
	if !ok {
		return false
	}
	if time.Now().After(until) {
		delete(t.lockouts, key)
		delete(t.attempts, key)
		delete(t.lastAttempt, key)
		return false
	}
	return true
}

func (t *LoginAttemptsTracker) RecordSuccess(key string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.attempts, key)
	delete(t.lockouts, key)
	delete(t.lastAttempt, key)
}

func (t *LoginAttemptsTracker) GetDelay(key string) time.Duration {
	t.mu.RLock()
	defer t.mu.RUnlock()
	count := t.attempts[key]
	if count <= 1 {
		return 0
	}
	delay := time.Duration(100*(1<<(count-1))) * time.Millisecond
	if delay > 2*time.Second {
		delay = 2 * time.Second
	}
	return delay
}
