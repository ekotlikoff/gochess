package httpserver

import (
	"gochess/internal/server/http/session/memoryprovider/memory"
	"gochess/internal/server/http/session/session"
)

var globalSessions *session.Manager

// initialize in init() function
func init() {
	globalSessions, _ = session.NewManager(memory.GetProvider(), "gosessionid", 3600)
	go globalSessions.GC()
}
