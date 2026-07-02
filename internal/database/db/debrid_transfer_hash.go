package db

import (
	"errors"
	"seanime/internal/database/models"

	"gorm.io/gorm"
)

func (db *Database) GetDebridTransferHashes(provider string) ([]*models.DebridTransferHash, error) {
	var res []*models.DebridTransferHash
	err := db.gormdb.Where("provider = ?", provider).Find(&res).Error
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (db *Database) UpsertDebridTransferHash(provider, transferId, hash string) error {
	var existing models.DebridTransferHash
	err := db.gormdb.Where("provider = ? AND transfer_id = ?", provider, transferId).First(&existing).Error
	if err == nil {
		existing.Hash = hash
		return db.gormdb.Save(&existing).Error
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	return db.gormdb.Create(&models.DebridTransferHash{
		Provider:   provider,
		TransferID: transferId,
		Hash:       hash,
	}).Error
}

func (db *Database) DeleteDebridTransferHash(provider, transferId string) error {
	return db.gormdb.Where("provider = ? AND transfer_id = ?", provider, transferId).Delete(&models.DebridTransferHash{}).Error
}
