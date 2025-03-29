package utils

import (
	"log"
)

func Contains(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}


// LogBoxedMessage is a simpler version that works well in all terminals
func BoxLog(message string) {
	log.Println(message)
}