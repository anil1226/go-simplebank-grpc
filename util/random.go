package util

import (
	"fmt"
	"math/rand"
	"strings"
)

// func init() {
// 	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
// 	rng.Seed(1)
// }

const alphabet = "abcdefghijklmnopqrstuvwxyz"

func RandomInt(min int64, max int64) int64 {
	return min + rand.Int63n(max-min+1)
}

func RandomString(n int) string {
	var sb strings.Builder
	k := len(alphabet)
	for range n {
		c := alphabet[rand.Intn(k)]
		sb.WriteByte(c)
	}
	return sb.String()
}

func RandomOwner() string {
	return RandomString(6)
}

func RandomMoney() int64 {
	return RandomInt(1, 100)
}

func RandomCurrency() string {
	currencies := []string{USD, INR}
	return currencies[rand.Intn(len(currencies))]
}

func RandomEmail() string {
	return fmt.Sprintf("%s@email.com", RandomString(6))
}
