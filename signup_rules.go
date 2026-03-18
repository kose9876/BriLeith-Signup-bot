package main

const maxSignupUsersPerDay = 8

func userHasSignupDay(store map[string]map[string][]string, weekKey string, userID string, day string) bool {
	if store[weekKey] == nil {
		return false
	}

	for _, existingDay := range store[weekKey][userID] {
		if existingDay == day {
			return true
		}
	}

	return false
}

func countSignupUsersForDay(store map[string]map[string][]string, weekKey string, day string) int {
	if store[weekKey] == nil {
		return 0
	}

	count := 0
	for _, days := range store[weekKey] {
		for _, existingDay := range days {
			if existingDay == day {
				count++
				break
			}
		}
	}

	return count
}

func isSignupDayFull(store map[string]map[string][]string, weekKey string, userID string, day string) bool {
	if userHasSignupDay(store, weekKey, userID, day) {
		return false
	}

	return countSignupUsersForDay(store, weekKey, day) >= maxSignupUsersPerDay
}
