package db

import (
	"os"

	postgrest "github.com/supabase/postgrest-go"
)

func CreateClient() (*postgrest.Client, error) {
	supabaseUrl := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_KEY")

	client := postgrest.NewClient(supabaseUrl+"/rest/v1", "", nil)

	if client.ClientError != nil {
		return nil, client.ClientError
	}

	client.TokenAuth(supabaseKey)

	return client, nil
}
