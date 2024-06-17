package webserver

import (
	"crypto/sha256"
	"fmt"
	"time"
)

// Hash returns the sha256 for  string
func Hash(key string) string {
	h := sha256.New()
	// hash.Hash.Write never returns an error.
	//nolint: errcheck
	h.Write([]byte(string(key)))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func timeNowToDateTimeFormatted() string {
	loc, _ := time.LoadLocation("Europe/Paris")
	current_time := time.Now().In(loc)
	return current_time.Format("Jan 2, 2006 at 3:04 PM")
}
