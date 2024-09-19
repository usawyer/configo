package helper

import "log"

func DefaultHandleError(err error) {
	log.Printf("ConfigManager error: %v", err)
}
