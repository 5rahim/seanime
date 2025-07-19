package core

import (
	"seanime/internal/util"
	"time"
)

// GetServerPasswordHMACAuth returns an HMAC authenticator using the hashed server password as the base secret
// This is used for server endpoints that don't use Nakama
func (a *App) GetServerPasswordHMACAuth() *util.HMACAuth {
	var secret string
	if a.Config != nil && a.Config.Server.Password != "" {
		secret = a.ServerPasswordHash
	} else {
		secret = "seanime-default-secret"
	}

	return util.NewHMACAuth(secret, 24*time.Hour)
}
