package db

type KVStore struct {
	ID        string `json:"id",omitempty`
	UserID    string `json:"user_id"`
	Key       string `json:"key"`
	Value     string `json:"value"`
	CreatedAt string `json:"created_at",omitempty`
}

type KVStoreInsert struct {
	UserID string `json:"user_id"`
	Key    string `json:"key"`
	Value  string `json:"value"`
}

func GetKV(userId string, key string) (*KVStore, error) {
	client, err := CreateClient()
	if err != nil {
		panic(err)
	}

	var result *KVStore

	_, err = client.From("kvs").
		Select("*", "exact", false).
		Single().
		Eq("user_id", userId).
		Eq("key", userId+"|"+key).
		ExecuteTo(&result)

	if err != nil || result == nil {
		return nil, err
	}

	result.Key = key

	return result, nil
}

func SetKV(userId string, key string, value string) error {
	client, err := CreateClient()
	if err != nil {
		return err
	}

	_, _, err = client.From("kvs").Insert(KVStoreInsert{
		UserID: userId,
		Key:    userId + "|" + key,
		Value:  value,
	}, true, "key", "", "exact").Execute()

	if err != nil {
		return err
	}

	return nil
}
