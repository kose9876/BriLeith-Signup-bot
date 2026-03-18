package main

import (
	"fmt"
	"time"
)

func getCurrentWeekKey() string {
	now := time.Now()
	weekday := int(now.Weekday())

	if weekday == 0 {
		weekday = 7
	}

	monday := now.AddDate(0, 0, -(weekday - 1))
	return monday.Format("2006-01-02")
}

func getSignupWeekKey() string {
	now := time.Now()
	weekday := int(now.Weekday())

	if weekday == 0 {
		weekday = 7
	}

	thisMonday := now.AddDate(0, 0, -(weekday - 1))
	nextMonday := thisMonday.AddDate(0, 0, 7)

	return nextMonday.Format("2006-01-02")
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
