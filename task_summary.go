package main

import "strings"

func buildBoss1TaskText(day DayAssignment) string {
	assignments := assignBoss1(day, map[string][]string{})
	return buildBoss1TaskTextFromAssignments(assignments)
}

func buildBoss1TaskTextFromAssignments(assignments []WorkAssignment) string {
	var lines []string
	lines = append(lines, "1王")

	for _, assignment := range assignments {
		lines = append(lines, assignment.Label+"："+assignment.Assignee)
	}

	return strings.Join(lines, "\n")
}

func buildTestFullDaySummaryText(weekKey string, dayKey string) string {
	assignment := buildWeekAssignmentFromStore(testWeeklySignups, weekKey)
	day := assignment.Days[dayKey]
	boss1Assignments := assignBoss1(day, map[string][]string{})
	boss2Assignments := assignBoss2Group(day, map[string][]string{})
	boss3Assignments := buildTestBoss3Assignments(weekKey, dayKey)

	return buildFullDaySummaryTextWithAssignments(weekKey, dayKey, day, boss1Assignments, boss2Assignments, boss3Assignments)
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
	assignment := buildWeekAssignmentFromStore(weeklySignups, weekKey)
	day := assignment.Days[dayKey]
	boss1Assignments := assignBoss1(day, map[string][]string{})
	boss2Assignments := assignBoss2Group(day, map[string][]string{})
	boss3Assignments := buildFormalBoss3Assignments(weekKey, dayKey)

	return buildFullDaySummaryTextWithAssignments(weekKey, dayKey, day, boss1Assignments, boss2Assignments, boss3Assignments)
}

func buildFullDaySummaryTextFromStore(signups map[string]map[string][]string, weekKey string, dayKey string) string {
	assignment := buildWeekAssignmentFromStore(signups, weekKey)
	day := assignment.Days[dayKey]
	boss1Assignments := assignBoss1(day, map[string][]string{})
	boss2Assignments := assignBoss2Group(day, map[string][]string{})
	boss3Assignments := assignBoss3(day, map[string][]string{})

	return buildFullDaySummaryTextWithAssignments(weekKey, dayKey, day, boss1Assignments, boss2Assignments, boss3Assignments)
}

func buildFullDaySummaryTextWithAssignments(weekKey string, dayKey string, day DayAssignment, boss1Assignments []WorkAssignment, boss2Assignments []GroupAssignment, boss3Assignments []WorkAssignment) string {
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

	boss1Text := buildBoss1TaskTextFromAssignments(boss1Assignments)
	boss2Text := buildBoss2TaskTextFromAssignments(boss2Assignments)
	boss3Text := buildBoss3TaskTextFromAssignments(boss3Assignments)

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
	return buildBoss2TaskTextFromAssignments(assignments)
}

func buildBoss2TaskTextFromAssignments(assignments []GroupAssignment) string {
	var lines []string
	lines = append(lines, "2王")

	for _, assignment := range assignments {
		lines = append(lines, assignment.Label+"："+strings.Join(assignment.Assignees, "、"))
	}

	return strings.Join(lines, "\n")
}

func buildBoss3TaskText(day DayAssignment) string {
	assignments := assignBoss3(day, map[string][]string{})
	return buildBoss3TaskTextFromAssignments(assignments)
}

func buildBoss3TaskTextFromAssignments(assignments []WorkAssignment) string {
	var lines []string
	lines = append(lines, "3王")

	for _, assignment := range assignments {
		lines = append(lines, assignment.Label+"："+formatWorkAssignmentAssignees(assignment))
	}

	return strings.Join(lines, "\n")
}

func formatWorkAssignmentAssignees(assignment WorkAssignment) string {
	names := []string{}
	if assignment.Assignee != "" {
		names = append(names, assignment.Assignee)
	}
	names = append(names, assignment.ExtraAssignees...)
	if len(names) == 0 {
		return "缺人"
	}
	return strings.Join(names, "、")
}
