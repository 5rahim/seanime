package db

import (
	"errors"
	"github.com/seanime-app/seanime/internal/database/models"
	"gorm.io/gorm/clause"
)

func (db *Database) UpsertAccount(acc *models.Account) (*models.Account, error) {
	err := db.gormdb.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		UpdateAll: true,
	}).Create(acc).Error

	if err != nil {
		db.Logger.Error().Err(err).Msg("Failed to save account in the database")
		return nil, err
	}
	return acc, nil
}

func (db *Database) GetAccount() (*models.Account, error) {
	var acc models.Account
	err := db.gormdb.Last(&acc).Error
	if err != nil {
		return nil, err
	}
	if acc.Username == "" || acc.Token == "" || acc.Viewer == nil {
		return nil, errors.New("account does not exist")
	}
	return &acc, err
}

// GetAnilistToken retrieves the AniList token from the account or returns an empty string
func (db *Database) GetAnilistToken() string {
	acc, err := db.GetAccount()
	if err != nil {
		return ""
	}
	return acc.Token
}
