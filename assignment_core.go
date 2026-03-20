package main

import (
	"sort"
	"strings"
)

func getOrderedDayKeys() []string {
	return []string{
		"day_mon",
		"day_tue",
		"day_wed",
		"day_thu",
		"day_fri",
		"day_sat",
		"day_sun",
	}
}

func collectDayApplicantsFromStore(signups map[string]map[string][]string, weekKey string) map[string][]string {
	result := map[string][]string{
		"day_mon": {},
		"day_tue": {},
		"day_wed": {},
		"day_thu": {},
		"day_fri": {},
		"day_sat": {},
		"day_sun": {},
	}

	users := signups[weekKey]
	for userID, days := range users {
		for _, day := range days {
			result[day] = append(result[day], userID)
		}
	}

	for dayKey := range result {
		sort.Slice(result[dayKey], func(i, j int) bool {
			leftID := result[dayKey][i]
			rightID := result[dayKey][j]

			leftKey := strings.ToLower(getDisplayName(leftID))
			rightKey := strings.ToLower(getDisplayName(rightID))
			if leftKey == rightKey {
				return leftID < rightID
			}
			return leftKey < rightKey
		})
	}

	return result
}

func collectDayApplicants(weekKey string) map[string][]string {
	return collectDayApplicantsFromStore(weeklySignups, weekKey)
}

func hasRole(userID string, role string, useSubRole bool) bool {
	profile, exists := userProfiles[userID]
	if !exists {
		return false
	}

	if useSubRole {
		return profile.SubRole == role
	}

	return profile.MainRole == role
}

func pickFirstByRole(candidates []string, role string, useSubRole bool) (string, []string) {
	for i, userID := range candidates {
		if hasRole(userID, role, useSubRole) {
			remaining := append([]string{}, candidates[:i]...)
			remaining = append(remaining, candidates[i+1:]...)
			return userID, remaining
		}
	}

	return "", candidates
}

func assignOneDay(applicants []string) DayAssignment {
	remaining := append([]string{}, applicants...)

	tank, remaining := pickFirstByRole(remaining, "tank", false)
	if tank == "" {
		tank, remaining = pickFirstByRole(remaining, "tank", true)
	}
	if tank == "" {
		tank = "缺坦"
	}

	healer, remaining := pickFirstByRole(remaining, "healer", false)
	if healer == "" {
		healer, remaining = pickFirstByRole(remaining, "healer", true)
	}
	if healer == "" {
		healer = "缺補"
	}

	return DayAssignment{
		Tank:   tank,
		Healer: healer,
		DPS:    remaining,
	}
}

func buildWeekAssignmentFromStore(signups map[string]map[string][]string, weekKey string) WeekAssignment {
	dayApplicants := collectDayApplicantsFromStore(signups, weekKey)

	result := WeekAssignment{
		Days: map[string]DayAssignment{},
	}

	for _, dayKey := range getOrderedDayKeys() {
		applicants := dayApplicants[dayKey]
		result.Days[dayKey] = assignOneDay(applicants)
	}

	return result
}

func buildWeekAssignment(weekKey string) WeekAssignment {
	return buildWeekAssignmentFromStore(weeklySignups, weekKey)
}

func getDisplayName(userID string) string {
	if userID == "缺坦" || userID == "缺補" || userID == "缺人" {
		return userID
	}

	profile, exists := userProfiles[userID]
	if exists {
		if profile.GameName != "" {
			return profile.GameName
		}
		if profile.DisplayName != "" {
			return profile.DisplayName
		}
	}

	return userID
}
