package api

import (
	"math/rand"
	"time"
)

const (
	randLength = 32
	allstring="abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ123456789"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func GetRandString() string {
	b := make([]byte, randLength)
    for i := range b {
    	b[i] = allstring[rand.Intn(len(allstring))]
	}
	return string(b)
}