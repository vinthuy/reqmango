//go:build ignore

package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func main() {
	// Never hardcode a signing key here: this helper must use the same SECRET_KEY
	// the server is configured with, or it would hand out forgeable tokens.
	secret := os.Getenv("SECRET_KEY")
	if secret == "" {
		log.Fatal("SECRET_KEY is not set; export the same value the server uses")
	}
	userID := uint64(4)
	expiresAt := time.Now().Add(24 * time.Hour)

	claims := jwt.MapClaims{
		"sub": strconv.FormatUint(userID, 10),
		"exp": expiresAt.Unix(),
		"iat": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(secret))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Bearer %s\n", tokenStr)
}
