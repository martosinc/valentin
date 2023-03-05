// package main

// import (
// 	"os"
// 	"strings"

// 	"github.com/joho/godotenv"
// )

// func main() {
// 	godotenv.Load(".env")
// 	data := os.Getenv("USERS")
// 	users := strings.Split(string(data), ",")
// 	userJob := make(map[string]int)
// 	for _, u := range users {
// 		userJob[u] = -1
// 	}
// }