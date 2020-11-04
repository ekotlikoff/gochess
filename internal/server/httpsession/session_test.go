package session

import (
	"net/http/httptest"
	"strings"
	"testing"
)

var globalSessions *Manager

// initialize in init() function
func init() {
	globalSessions, _ = NewManager("memory", "gosessionid", 3600)
	go globalSessions.GC()
}

func TestSession(t *testing.T) {
	w := httptest.ResponseRecorder{}
	r := httptest.NewRequest("get", "/", strings.NewReader("Hello Reader"))
	sess := globalSessions.SessionStart(w, &r)
	t.Error()
}
