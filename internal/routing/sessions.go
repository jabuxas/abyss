package routing

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

type Session struct {
	Filename  string
	CreatedAt time.Time
}

var SessionStore = struct {
	sync.RWMutex
	data map[string]Session
}{data: make(map[string]Session)}

const sessionCookieName = "abyss_session"
const sessionDuration = time.Hour * 1

func NewSession(filename string) (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	token := hex.EncodeToString(b)

	SessionStore.Lock()
	defer SessionStore.Unlock()

	SessionStore.data[token] = Session{
		Filename:  filename,
		CreatedAt: time.Now(),
	}

	return token, nil
}

func GetSession(token, filename string) bool {
	SessionStore.RLock()
	defer SessionStore.RUnlock()

	session, exists := SessionStore.data[token]
	if !exists {
		return false
	}

	return session.Filename == filename && time.Since(session.CreatedAt) < sessionDuration
}

func init() {
	go func() {
		for range time.Tick(10 * time.Minute) {
			SessionStore.Lock()
			for token, session := range SessionStore.data {
				if time.Since(session.CreatedAt) > sessionDuration {
					delete(SessionStore.data, token)
				}
			}
			SessionStore.Unlock()
		}
	}()
}
