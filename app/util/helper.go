package util

import (
	"log"
	"os"
)

// Debug logger function
func DebugLog(title string, message ...interface{}) {
	if os.Getenv("DEBUG") == "true" {
		log.Println("DEBUG:", title, message)
	}
}
