package db

type KVStore struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	Key       string `json:"key"`
	Value     string `json:"value"`
	CreatedAt string `json:"created_at"`
}

func (KVStore) TableName() string {
	return "kvs"
}

func GetKV(userId string, key string) (*KVStore, error) {
	var result *KVStore

	err := DB.First(&result, "user_id = ? AND key = ?", userId, userId+"|"+key).Error

	if err != nil || result == nil {
		return nil, err
	}

	result.Key = key

	return result, nil
}

func SetKV(userId string, key string, value string) error {
	err := DB.Exec(
		"INSERT INTO kvs (user_id, key, value) VALUES (?, ?, ?) ON CONFLICT (key) DO UPDATE SET value = ?",
		userId,
		userId+"|"+key,
		value,
		value,
	).Error
	if err != nil {
		return err
	}

	return nil
}
