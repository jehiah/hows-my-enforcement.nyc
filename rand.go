package main

import (
	"crypto/rand"
	"encoding/base64"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"
)

const (
	randBytes = 7
)

var pool = sync.Pool{New: func() interface{} { return make([]byte, randBytes) }}

var cleaner = strings.NewReplacer("-", "", "_", "")

// Randomness returns some random string values
// This function will panic if unable to read from crypto/rand
func Randomness() (s string) {
	var randomness string
	for i := 0; i < 5; i++ {
		b := pool.Get().([]byte)
		_, err := rand.Read(b)
		if err != nil {
			log.Panicf("ERROR: failed to read %d random bytes for shortid: %s", randBytes, err)
		}
		randomness = cleaner.Replace(base64.RawURLEncoding.EncodeToString(b))
		pool.Put(b)
		if len(randomness) >= 6 {
			randomness = randomness[:6]
		}
		if len(randomness) == 6 {
			break
		}
	}
	if len(randomness) != 6 {
		log.Panic("ERROR: failed generating shortid")
	}

	return randomness
}
