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

// CanonicalTokens returns deduplicated sorted tokens for deterministic name matching.
func CanonicalTokens(raw string) []string {
	normalized := NormalizeText(raw)
	if normalized == "" {
		return nil
	}
	parts := strings.Fields(normalized)
	set := make(map[string]struct{}, len(parts))
	for _, p := range parts {
		if p == "" {
			continue
		}
		set[p] = struct{}{}
	}
	out := make([]string, 0, len(set))
	for t := range set {
		out = append(out, t)
	}
	sort.Strings(out)
	return out
}

// TokenOverlapScore computes a deterministic Jaccard overlap score [0,1].
func TokenOverlapScore(left, right []string) float64 {
	if len(left) == 0 || len(right) == 0 {
		return 0
	}
	ls := make(map[string]struct{}, len(left))
	rs := make(map[string]struct{}, len(right))
	for _, t := range left {
		ls[t] = struct{}{}
	}
	for _, t := range right {
		rs[t] = struct{}{}
	}
	intersect := 0
	for t := range ls {
		if _, ok := rs[t]; ok {
			intersect++
		}
	}
	union := len(ls)
	for t := range rs {
		if _, ok := ls[t]; !ok {
			union++
		}
	}
	if union == 0 {
		return 0
	}
	return float64(intersect) / float64(union)
}
