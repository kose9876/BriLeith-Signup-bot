package main

import "strings"

func buildBoss1TaskText(day DayAssignment) string {
	assignments := assignBoss1(day, map[string][]string{})

	var lines []string
	lines = append(lines, "1王")

	for _, assignment := range assignments {
		lines = append(lines, assignment.Label+"："+assignment.Assignee)
	}

	return strings.Join(lines, "\n")
}

func buildDayBossSummaryText(weekKey string, dayKey string) string {
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

	dayDateText := getDayDateText(weekKey, dayKey)

	return getWeekRangeText(weekKey) + " " + dayNames[dayKey] + "（" + dayDateText + "）Boss分配\n\n" +
		buildBoss1TaskText(day)
}

func buildFullDaySummaryText(weekKey string, dayKey string) string {
	return buildFullDaySummaryTextFromStore(weeklySignups, weekKey, dayKey)
}

func buildFullDaySummaryTextFromStore(signups map[string]map[string][]string, weekKey string, dayKey string) string {
	assignment := buildWeekAssignmentFromStore(signups, weekKey)
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

	dayDateText := getDayDateText(weekKey, dayKey)

	var dpsNames []string
	for _, userID := range day.DPS {
		dpsNames = append(dpsNames, getDisplayName(userID))
	}

	dpsText := "無"
	if len(dpsNames) > 0 {
		dpsText = strings.Join(dpsNames, "、")
	}

	boss1Text := buildBoss1TaskText(day)
	boss2Text := buildBoss2TaskText(day)
	boss3Text := buildBoss3TaskText(day)

	return getWeekRangeText(weekKey) + " " + dayNames[dayKey] + "（" + dayDateText + "）分配總覽\n\n" +
		"隊伍配置\n" +
		"坦克：" + getDisplayName(day.Tank) + "\n" +
		"補師：" + getDisplayName(day.Healer) + "\n" +
		"輸出：" + dpsText + "\n\n" +
		boss1Text + "\n\n" +
		boss2Text + "\n\n" +
		boss3Text
}

func buildBoss2TaskText(day DayAssignment) string {
	assignments := assignBoss2Group(day, map[string][]string{})

	var lines []string
	lines = append(lines, "2王")

	for _, assignment := range assignments {
		lines = append(lines, assignment.Label+"："+strings.Join(assignment.Assignees, "、"))
	}

	return strings.Join(lines, "\n")
}

func buildBoss3TaskText(day DayAssignment) string {
	assignments := assignBoss3(day, map[string][]string{})

	var lines []string
	lines = append(lines, "3王")

	for _, assignment := range assignments {
		lines = append(lines, assignment.Label+"："+assignment.Assignee)
	}

	return strings.Join(lines, "\n")
}
