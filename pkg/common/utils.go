package common

import (
	"log"
	"strings"
	"time"
)

func ExtractNumericID(fullID string) string {
	parts := strings.Split(fullID, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return fullID
}

func CalculateDaysUntilExpire(expireDate string) int {
	parsedExpireDate, err := time.Parse("2006-01-02", expireDate)
	if err != nil {
		log.Printf("error to parse time: %s\n", err)
		return 0
	}
	days := int(time.Until(parsedExpireDate).Hours() / 24)
	if days < 0 {
		days = 0
	}
	return days
}
