package meal

import (
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"strings"
)

var (
	multiSpaceRegexp = regexp.MustCompile(`\s+`)
	multiCommaRegexp = regexp.MustCompile(`,+`)
)

// NormalizeText normalizes raw meal text for stable fingerprinting.
func NormalizeText(raw string) string {
	normalized := strings.ToLower(strings.TrimSpace(raw))
	normalized = strings.ReplaceAll(normalized, "，", ",")
	normalized = multiCommaRegexp.ReplaceAllString(normalized, ",")
	normalized = strings.Trim(normalized, ",")
	normalized = strings.ReplaceAll(normalized, ",", " ")
	normalized = multiSpaceRegexp.ReplaceAllString(normalized, " ")
	return strings.TrimSpace(normalized)
}

// FingerprintFromText returns a stable SHA-256 fingerprint in hex format.
func FingerprintFromText(raw string) string {
	normalized := NormalizeText(raw)
	sum := sha256.Sum256([]byte(normalized))
	return hex.EncodeToString(sum[:])
}
