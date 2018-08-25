package webService

import (
	"runtime"
	"sync"
	"time"

	"github.com/dchest/uniuri"
)

const cleanupAfter = 20 * time.Minute

type tokenEntry struct {
	value string
	ts    int64 // unix nano
}

type simpleXSRFheckInternal struct {
	tokens      map[string]tokenEntry
	lock        sync.RWMutex
	stopCleaner chan bool
}

// SimpleXSRFCheck creates/compares and stores the XSRF tokens for the associated ip. After cleanupAfter amount of time
// a cleanup go routine removes old entries.
type SimpleXSRFCheck struct {
	// the real struct. The outer hull is a wrapper that allows the 'SetFinalizer' function to work probably.
	*simpleXSRFheckInternal
}

// NewSimpleXSRFCheck creates a new instance
func NewSimpleXSRFCheck() *SimpleXSRFCheck {
	obj := simpleXSRFheckInternal{tokens: make(map[string]tokenEntry), stopCleaner: make(chan bool)}
	go obj.cleanup()

	// we need this wrapper or the SetFinalizer would never call
	wraper := &SimpleXSRFCheck{&obj}
	runtime.SetFinalizer(wraper, stopCleaner)
	return wraper
}

func stopCleaner(sx *SimpleXSRFCheck) {
	select {
	case sx.stopCleaner <- true:
		break
		// don't block!
	default:
		break
	}
}

func (sx *simpleXSRFheckInternal) NewToken(ip string) string {
	token := tokenEntry{value: uniuri.NewLen(10), ts: time.Now().UnixNano()}
	sx.lock.Lock()
	sx.tokens[ip] = token
	sx.lock.Unlock()
	return token.value
}

// CheckAndClearToken returns true if the given token matches for the given ip. If no entry was found for the ip, false is returned.
// At the end the ip entry is removed from the internal map.
func (sx *simpleXSRFheckInternal) CheckAndClearToken(ip string, token string) bool {
	sx.lock.RLock()
	defer sx.lock.RUnlock()
	savedToken, ok := sx.tokens[ip]
	if !ok {
		return false
	}
	delete(sx.tokens, ip)
	return savedToken.value == token
}

func (sx *simpleXSRFheckInternal) cleanup() {
	ticker := time.NewTicker(cleanupAfter)

	for {
		select {
		case <-sx.stopCleaner:
			ticker.Stop()
			return
		case <-ticker.C:
			cleanupValue := time.Now().UnixNano() - int64(cleanupAfter)
			sx.lock.Lock()
			for k, v := range sx.tokens {
				if v.ts < cleanupValue {
					delete(sx.tokens, k)
				}
			}
			sx.lock.Unlock()
		}
	}
}
