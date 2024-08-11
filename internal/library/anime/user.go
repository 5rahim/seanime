package anime

import (
	"errors"
	"github.com/goccy/go-json"
	"seanime/internal/api/anilist"
	"seanime/internal/database/models"
)

type User struct {
	Viewer *anilist.GetViewer_Viewer `json:"viewer"`
	Token  string                    `json:"token"`
}

// NewUser creates a new User entity from a models.User
// This is returned to the client
func NewUser(model *models.Account) (*User, error) {
	if model == nil {
		return nil, errors.New("account is nil")
	}
	var acc anilist.GetViewer_Viewer
	if err := json.Unmarshal(model.Viewer, &acc); err != nil {
		return nil, err
	}
	return &User{
		Viewer: &acc,
		Token:  model.Token,
	}, nil
}
