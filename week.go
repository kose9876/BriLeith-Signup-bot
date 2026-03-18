package main

import (
	"fmt"
	"time"
)

func getCurrentWeekKey() string {
	return getCurrentWeekKeyAt(nowInBotLocation())
}

func getSignupWeekKey() string {
	return getSignupWeekKeyAt(nowInBotLocation())
}

func getWeekRangeText(weekKey string) string {
	start, err := time.Parse("2006-01-02", weekKey)
	if err != nil {
		return weekKey
	}

	end := start.AddDate(0, 0, 6)

	return fmt.Sprintf("%d/%d~%d/%d",
		start.Month(), start.Day(),
		end.Month(), end.Day(),
	)
}
