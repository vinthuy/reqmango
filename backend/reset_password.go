package main

import (
	"fmt"
	"log"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	hash, err := bcrypt.GenerateFromPassword([]byte("demo1234"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal(err)
	}

	db, err := gorm.Open(postgres.Open("postgres://postgres:postgres@localhost:5432/reqmango"), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	result := db.Exec("UPDATE users SET password_hash = ?", string(hash))
	if result.Error != nil {
		log.Fatal(result.Error)
	}

	fmt.Printf("Updated %d users with new password hash\n", result.RowsAffected)
}
