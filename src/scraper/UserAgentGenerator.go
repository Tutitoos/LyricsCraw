package scraper

import (
	crypto_rand "crypto/rand"
	"math/big"
)

// userAgents is the default list of User-Agent strings.
var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 6.1; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.6312.86 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.6261.94 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 13_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.6367.91 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 12_6_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.6312.105 Safari/537.36",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.6261.70 Safari/537.36",
	"Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:123.0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Linux; Android 13; Pixel 7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.6367.61 Mobile Safari/537.36",
	"Mozilla/5.0 (Linux; Android 12; SM-G991B) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.6312.86 Mobile Safari/537.36",
	"Mozilla/5.0 (iPhone; CPU iPhone OS 16_5 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) CriOS/124.0.6367.61 Mobile/15E148 Safari/604.1",
}

// GetRandomUserAgent returns a random User-Agent string from the default list.
// Uses crypto/rand for better randomness.
func GetRandomUserAgent() string {
	return GetRandomUserAgentFromList(userAgents)
}

// GetRandomUserAgentFromList returns a random User-Agent from a custom list.
// Returns empty string if the list is empty.
func GetRandomUserAgentFromList(list []string) string {
	if len(list) == 0 {
		return ""
	}
	max := big.NewInt(int64(len(list)))
	n, err := crypto_rand.Int(crypto_rand.Reader, max)
	if err != nil {
		return list[0] // fallback to first if random fails
	}
	return list[n.Int64()]
}
