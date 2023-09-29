package db

import (
	"fmt"
	"os"

	postgrest "github.com/supabase/postgrest-go"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func createDB() gorm.DB {
	db, err := gorm.Open(postgres.Open(os.Getenv("POSTGRES_URI")), &gorm.Config{})
	if err != nil {
		fmt.Println("POSTGRES_URI: ", os.Getenv("POSTGRES_URI"))
		panic("POSTGRES_URI: " + os.Getenv("POSTGRES_URI"))
	}

	return *db
}

var DB gorm.DB = createDB()

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
