package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

type DayAssignment struct {
	Tank   string
	Healer string
	DPS    []string
}

type WeekAssignment struct {
	Days map[string]DayAssignment
}

type DayTaskAssignments struct {
	Boss1 []WorkAssignment
	Boss2 []GroupAssignment
	Boss3 []WorkAssignment
}

type WeekTaskAssignments struct {
	Days map[string]DayTaskAssignments
}

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

func buildDayAssignmentText(weekKey string, dayKey string) string {
	assignment := buildWeekAssignment(weekKey)
	day := assignment.Days[dayKey]

	dayNames := map[string]string{
		"day_mon": "周一",
		"day_tue": "周二",
		"day_wed": "周三",
		"day_thu": "周四",
		"day_fri": "周五",
		"day_sat": "周六",
		"day_sun": "周日",
	}

	var dpsNames []string
	for _, userID := range day.DPS {
		dpsNames = append(dpsNames, getDisplayName(userID))
	}

	dpsText := "無"
	if len(dpsNames) > 0 {
		dpsText = strings.Join(dpsNames, "、")
	}

	dayDateText := getDayDateText(weekKey, dayKey)
	return getWeekRangeText(weekKey) + " " + dayNames[dayKey] + "（" + dayDateText + "）職位分配\n\n" +

		"坦克：" + getDisplayName(day.Tank) + "\n" +
		"補師：" + getDisplayName(day.Healer) + "\n" +
		"輸出：" + dpsText
}

func buildSummaryComponents() []discordgo.MessageComponent {
	return []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{Label: "周一", Style: discordgo.PrimaryButton, CustomID: "summary_day_mon"},
				discordgo.Button{Label: "周二", Style: discordgo.PrimaryButton, CustomID: "summary_day_tue"},
				discordgo.Button{Label: "周三", Style: discordgo.PrimaryButton, CustomID: "summary_day_wed"},
				discordgo.Button{Label: "周四", Style: discordgo.PrimaryButton, CustomID: "summary_day_thu"},
				discordgo.Button{Label: "周五", Style: discordgo.PrimaryButton, CustomID: "summary_day_fri"},
			},
		},
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{Label: "周六", Style: discordgo.SuccessButton, CustomID: "summary_day_sat"},
				discordgo.Button{Label: "周日", Style: discordgo.SuccessButton, CustomID: "summary_day_sun"},
			},
		},
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

func buildWeekAssignmentTextFromStore(signups map[string]map[string][]string, weekKey string) string {
	assignment := buildWeekAssignmentFromStore(signups, weekKey)

	dayNames := map[string]string{
		"day_mon": "周一",
		"day_tue": "周二",
		"day_wed": "周三",
		"day_thu": "周四",
		"day_fri": "周五",
		"day_sat": "周六",
		"day_sun": "周日",
	}

	dayOrder := []string{
		"day_mon",
		"day_tue",
		"day_wed",
		"day_thu",
		"day_fri",
		"day_sat",
		"day_sun",
	}

	var lines []string
	lines = append(lines, getWeekRangeText(weekKey)+" 職位分配")
	lines = append(lines, "")

	for _, dayKey := range dayOrder {
		day := assignment.Days[dayKey]

		var dpsNames []string
		for _, userID := range day.DPS {
			dpsNames = append(dpsNames, getDisplayName(userID))
		}

		dpsText := "無"
		if len(dpsNames) > 0 {
			dpsText = strings.Join(dpsNames, "、")
		}

		lines = append(lines, dayNames[dayKey])
		lines = append(lines, "坦克："+getDisplayName(day.Tank))
		lines = append(lines, "補師："+getDisplayName(day.Healer))
		lines = append(lines, "輸出："+dpsText)
		lines = append(lines, "")
	}

	return strings.Join(lines, "\n")
}

func buildWeekAssignmentText(weekKey string) string {
	return buildWeekAssignmentTextFromStore(weeklySignups, weekKey)
}
func getDayDateText(weekKey string, dayKey string) string {
	start, err := time.Parse("2006-01-02", weekKey)
	if err != nil {
		return ""
	}

	dayOffsets := map[string]int{
		"day_mon": 0,
		"day_tue": 1,
		"day_wed": 2,
		"day_thu": 3,
		"day_fri": 4,
		"day_sat": 5,
		"day_sun": 6,
	}

	offset, exists := dayOffsets[dayKey]
	if !exists {
		return ""
	}

	targetDate := start.AddDate(0, 0, offset)
	return fmt.Sprintf("%d/%d", targetDate.Month(), targetDate.Day())
}
