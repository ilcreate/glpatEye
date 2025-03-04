package common

import (
	"fmt"
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

func CalculateDaysUntilExpire(expireDate string) (int, error) {
	parsedExpireDate, err := time.Parse("2006-01-02", expireDate)
	if err != nil {
		return 0, fmt.Errorf("error to parse time: %s\n", err)
	}
	days := int(time.Until(parsedExpireDate).Hours() / 24)
	if days < 0 {
		return -1, fmt.Errorf("error calculate days before expire. quantity of days before expire is less than is.")
	}
	return days, nil
}
