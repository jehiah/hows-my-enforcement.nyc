package dthash

import (
	"time"
)

// These alphabets are in byte order from lowest byte to highest byte
const b36alphabet = "0123456789abcdefghijklmnopqrstuvwxyz"
const b62alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// DateHash is a 6 character string with Year, month, day, hour, minute, second
type DateHash time.Time

func (d DateHash) String() string {
	t := time.Time(d)
	year, month, day := t.Date()
	hr, min, sec := t.Clock()
	var dt = []byte{
		'0',
		b36alphabet[month],
		b36alphabet[day],
		b36alphabet[hr],
		b62alphabet[min],
		b62alphabet[sec],
	}
	if year >= 2000 && year < 2035 {
		dt[0] = b36alphabet[year%100]
	}
	return string(dt)
}
