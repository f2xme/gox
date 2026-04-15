package security

import "regexp"

var xssPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)<script`),
	regexp.MustCompile(`(?i)javascript:`),
	regexp.MustCompile(`(?i)onerror\s*=`),
	regexp.MustCompile(`(?i)onclick\s*=`),
}

var sqlInjectionPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)'\s*or\s*'1'='1`),
	regexp.MustCompile(`(?i)union\s+select`),
	regexp.MustCompile(`(?i)drop\s+table`),
	regexp.MustCompile(`--`),
}

func containsXSSPayload(value string) bool {
	for _, pattern := range xssPatterns {
		if pattern.MatchString(value) {
			return true
		}
	}
	return false
}

func containsSQLInjectionPayload(value string) bool {
	for _, pattern := range sqlInjectionPatterns {
		if pattern.MatchString(value) {
			return true
		}
	}
	return false
}
