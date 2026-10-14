//go:build ignore

package main

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	h, err := bcrypt.GenerateFromPassword([]byte("Admin@123"), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	fmt.Print(string(h))
}
