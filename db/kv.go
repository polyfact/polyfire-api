package db

import (
	"gorm.io/gorm/clause"
)

type KVStore struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	Key       string `json:"key"`
	Value     string `json:"value"`
	CreatedAt string `json:"created_at"`
}

type KVStoreInsert struct {
	UserID string `json:"user_id"`
	Key    string `json:"key"`
	Value  string `json:"value"`
}

func (KVStore) TableName() string {
	return "kvs"
}

func (KVStoreInsert) TableName() string {
	return "kvs"
}

func createCombinedKey(userID, key string) string {
	return userID + "|" + key
}

func SetKV(userID, key, value string) error {
	combinedKey := createCombinedKey(userID, key)

	kv := KVStoreInsert{Key: combinedKey, UserID: userID, Value: value}

	return DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},
		DoUpdates: clause.Assignments(map[string]interface{}{"value": value, "user_id": userID}),
	}).Create(&kv).Error
}

func GetKV(userID, key string) (*KVStore, error) {
	var result KVStore
	err := DB.First(&result, "key = ?", createCombinedKey(userID, key)).Error
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func DeleteKV(userID, key string) error {
	combinedKey := createCombinedKey(userID, key)

	kv := KVStore{Key: combinedKey}

	return DB.Where("key = ?", combinedKey).Delete(&kv).Error
}

func ListKV(userID string) ([]KVStore, error) {
	var results []KVStore
	err := DB.Where("user_id = ?", userID).Find(&results).Error
	if err != nil {
		return nil, err
	}
	return results, nil
}
