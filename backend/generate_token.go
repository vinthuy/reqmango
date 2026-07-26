//go:build ignore

package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func main() {
	secret := "change-me-in-production-use-a-long-random-string"
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