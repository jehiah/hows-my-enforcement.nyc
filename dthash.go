package main

import (
	"time"
)

// These alphabets are in byte order from lowest byte to highest byte
const b36alphabet = "0123456789abcdefghijklmnopqrstuvwxyz"

// DateHash is a 3 character string with Year, month and day
type DateHash time.Time

func (d DateHash) String() string {
	t := time.Time(d)
	year, month, day := t.Date()
	var dt = []byte{
		'0',
		b36alphabet[month],
		b36alphabet[day],
	}
	if year >= 2000 && year < 2035 {
		dt[0] = b36alphabet[year%100]
	}
	return string(dt)
}
