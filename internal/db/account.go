package db

import (
	"github.com/seanime-app/seanime-server/internal/models"
	"gorm.io/gorm/clause"
)

func (db *Database) UpsertAccount(acc *models.Account) (*models.Account, error) {
	err := db.gormdb.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		UpdateAll: true,
	}).Create(acc).Error

	if err != nil {
		db.logger.Error().Err(err).Msg("Failed to save account in the database")
		return nil, err
	}
	return acc, nil

}

func (db *Database) GetAccount() (*models.Account, error) {
	var acc models.Account
	err := db.gormdb.Last(&acc).Error
	if err != nil {
		db.logger.Error().Err(err).Msg("failed to get acc")
		return nil, err
	}
	return &acc, err
}
