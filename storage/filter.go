package storage

import "strings"

// NormalizeName normalizes a host name or search query by replacing hyphens
// with spaces, converting to lowercase, and collapsing multiple spaces.
func NormalizeName(s string) string {
	s = strings.ReplaceAll(s, "-", " ")
	s = strings.ToLower(s)
	return strings.Join(strings.Fields(s), " ")
}

// FindHostsByName searches for hosts whose normalized name contains the normalized query.
func FindHostsByName(hosts []Host, query string) []Host {
	normalizedQuery := NormalizeName(query)
	if normalizedQuery == "" {
		return nil
	}

	var matches []Host
	for _, host := range hosts {
		if strings.Contains(NormalizeName(host.Name), normalizedQuery) {
			matches = append(matches, host)
		}
	}
	return matches
}
