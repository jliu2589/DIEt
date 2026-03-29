package meal

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"diet/internal/models"
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

// FingerprintFromCanonicalStructure returns a stable fingerprint from canonical meal name + normalized item lines.
func FingerprintFromCanonicalStructure(canonicalName string, items []models.MealItem) string {
	base := NormalizeText(canonicalName)
	if base == "" {
		return ""
	}
	lines := make([]string, 0, len(items))
	for _, item := range items {
		if strings.TrimSpace(item.Name) == "" || item.Quantity == nil {
			continue
		}
		lines = append(lines, fmt.Sprintf("%s|%.4f|%s", NormalizeText(item.Name), *item.Quantity, strings.ToLower(strings.TrimSpace(item.Unit))))
	}
	sort.Strings(lines)
	payload := base + "::" + strings.Join(lines, ";")
	sum := sha256.Sum256([]byte(payload))
	return hex.EncodeToString(sum[:])
}
