package agent

import (
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"strings"
)

var (
	nonAlpha  = regexp.MustCompile(`[^\p{L}\p{N}\s]`)
	multiWS   = regexp.MustCompile(`\s+`)
	stopwords = map[string]bool{
		"the": true, "a": true, "an": true,
		"live": true, "tour": true, "concert": true,
		"feat": true, "featuring": true,
	}
)

func normalize(s string) string {
	s = strings.ToLower(s)
	s = nonAlpha.ReplaceAllString(s, " ")
	s = multiWS.ReplaceAllString(s, " ")
	return strings.TrimSpace(s)
}

func normalizeStripStopwords(s string) string {
	fields := strings.Fields(normalize(s))
	out := fields[:0]
	for _, f := range fields {
		if !stopwords[f] {
			out = append(out, f)
		}
	}
	return strings.Join(out, " ")
}

func sha(parts ...string) string {
	h := sha256.Sum256([]byte(strings.Join(parts, "|")))
	return hex.EncodeToString(h[:])
}

func EventFingerprint(title, dateISO, venue string) string {
	return sha(normalizeStripStopwords(title), dateISO, normalizeStripStopwords(venue))
}

func NewsFingerprint(headline string) string {
	return sha(normalizeStripStopwords(headline))
}
