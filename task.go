package main

import "strings"

type WorkAssignment struct {
	Label    string
	Assignee string
	UserID   string
}

type GroupAssignment struct {
	Label     string
	Assignees []string
	UserIDs   []string
}

func containsUser(users []string, target string) bool {
	for _, userID := range users {
		if userID == target {
			return true
		}
	}
	return false
}

func pickPreferredUser(taskKey string, users []string, history map[string][]string, match func(string) bool) (string, []string) {
	past := history[taskKey]

	for i := len(past) - 1; i >= 0; i-- {
		userID := past[i]
		if containsUser(users, userID) && match(userID) {
			return userID, removeUser(users, userID)
		}
	}

	return pickFirstMatching(users, match)
}

func pickPreferredAny(taskKey string, users []string, history map[string][]string) (string, []string) {
	return pickPreferredUser(taskKey, users, history, func(string) bool { return true })
}

func recordWorkAssignments(history map[string][]string, assignments []WorkAssignment) {
	for _, assignment := range assignments {
		if assignment.UserID == "" {
			continue
		}
		history[assignment.Label] = append(history[assignment.Label], assignment.UserID)
	}
}

func recordHistory(history map[string][]string, taskKey string, userID string) {
	if userID == "" {
		return
	}
	history[taskKey] = append(history[taskKey], userID)
}

func isExperienced(userID string) bool {
	profile, exists := userProfiles[userID]
	return exists && profile.Group == "experienced"
}
func hasCape(userID string) bool {
	profile, exists := userProfiles[userID]
	return exists && profile.HasCape
}

func removeUser(users []string, target string) []string {
	result := []string{}
	for _, userID := range users {
		if userID != target {
			result = append(result, userID)
		}
	}
	return result
}

func pickFirstMatching(users []string, match func(string) bool) (string, []string) {
	for i, userID := range users {
		if match(userID) {
			remaining := append([]string{}, users[:i]...)
			remaining = append(remaining, users[i+1:]...)
			return userID, remaining
		}
	}

	return "", users
}

func isNewbie(userID string) bool {
	profile, exists := userProfiles[userID]
	return exists && profile.Group == "newbie"
}

func assignBoss1(day DayAssignment, history map[string][]string) []WorkAssignment {
	assignments := []WorkAssignment{}
	remaining := append([]string{}, day.DPS...)

	positions := map[string]string{
		"boss1_1":  "缺人",
		"boss1_5":  "缺人",
		"boss1_7":  "缺人",
		"boss1_11": "缺人",
	}
	positionUsers := map[string]string{
		"boss1_1":  "",
		"boss1_5":  "",
		"boss1_7":  "",
		"boss1_11": "",
	}

	userID, rest := pickPreferredUser("boss1_11", remaining, history, func(id string) bool {
		return isExperienced(id) && hasCape(id)
	})
	if userID != "" {
		positions["boss1_11"] = getDisplayName(userID)
		positionUsers["boss1_11"] = userID
		remaining = rest
	}

	otherPositions := []string{"boss1_1", "boss1_5", "boss1_7"}
	for _, position := range otherPositions {
		userID, rest := pickPreferredUser(position, remaining, history, isExperienced)
		if userID != "" {
			positions[position] = getDisplayName(userID)
			positionUsers[position] = userID
			remaining = rest
		}
	}

	if positionUsers["boss1_11"] == "" && len(remaining) > 0 {
		userID, rest := pickPreferredAny("boss1_11", remaining, history)
		positions["boss1_11"] = getDisplayName(userID)
		positionUsers["boss1_11"] = userID
		remaining = rest
	}

	for _, position := range otherPositions {
		if positionUsers[position] == "" && len(remaining) > 0 {
			userID, rest := pickPreferredAny(position, remaining, history)
			positions[position] = getDisplayName(userID)
			positionUsers[position] = userID
			remaining = rest
		}
	}

	assignments = append(assignments, WorkAssignment{Label: "1點", Assignee: positions["boss1_1"], UserID: positionUsers["boss1_1"]})
	assignments = append(assignments, WorkAssignment{Label: "5點", Assignee: positions["boss1_5"], UserID: positionUsers["boss1_5"]})
	assignments = append(assignments, WorkAssignment{Label: "7點", Assignee: positions["boss1_7"], UserID: positionUsers["boss1_7"]})
	assignments = append(assignments, WorkAssignment{Label: "11點", Assignee: positions["boss1_11"], UserID: positionUsers["boss1_11"]})

	return assignments
}

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

func fillToTwo(assignees []string) []string {
	for len(assignees) < 2 {
		assignees = append(assignees, "缺人")
	}
	return assignees
}

func assignBoss2Group(day DayAssignment, history map[string][]string) []GroupAssignment {
	remaining := append([]string{}, day.DPS...)

	point3 := []string{}
	point6 := []string{}
	point9 := []string{}
	point12 := []string{}

	point3Users := []string{}
	point6Users := []string{}
	point9Users := []string{}
	point12Users := []string{}

	if day.Tank == "缺坦" {
		point3 = append(point3, "缺坦")
	} else {
		point3 = append(point3, getDisplayName(day.Tank))
		point3Users = append(point3Users, day.Tank)
	}

	userID, rest := pickPreferredUser("boss2_point3_helper", remaining, history, hasCape)
	if userID != "" {
		point3 = append(point3, getDisplayName(userID))
		point3Users = append(point3Users, userID)
		remaining = rest
	}

	if day.Healer == "缺補" {
		point6 = append(point6, "缺補")
	} else {
		point6 = append(point6, getDisplayName(day.Healer))
		point6Users = append(point6Users, day.Healer)
	}

	if len(remaining) > 0 {
		userID, rest := pickPreferredAny("boss2_point6_helper", remaining, history)
		point6 = append(point6, getDisplayName(userID))
		point6Users = append(point6Users, userID)
		remaining = rest
	}

	userID, rest = pickPreferredUser("boss2_point9_a", remaining, history, isNewbie)
	if userID != "" {
		point9 = append(point9, getDisplayName(userID))
		point9Users = append(point9Users, userID)
		remaining = rest
	} else if len(remaining) > 0 {
		userID, rest = pickPreferredAny("boss2_point9_a", remaining, history)
		point9 = append(point9, getDisplayName(userID))
		point9Users = append(point9Users, userID)
		remaining = rest
	}

	if len(remaining) > 0 {
		userID, rest = pickPreferredAny("boss2_point9_b", remaining, history)
		point9 = append(point9, getDisplayName(userID))
		point9Users = append(point9Users, userID)
		remaining = rest
	}

	if len(remaining) > 0 {
		userID, rest = pickPreferredAny("boss2_point12_a", remaining, history)
		point12 = append(point12, getDisplayName(userID))
		point12Users = append(point12Users, userID)
		remaining = rest
	}

	if len(remaining) > 0 {
		userID, rest = pickPreferredAny("boss2_point12_b", remaining, history)
		point12 = append(point12, getDisplayName(userID))
		point12Users = append(point12Users, userID)
		remaining = rest
	}

	point3 = fillToTwo(point3)
	point6 = fillToTwo(point6)
	point9 = fillToTwo(point9)
	point12 = fillToTwo(point12)

	return []GroupAssignment{
		{Label: "3點", Assignees: point3, UserIDs: point3Users},
		{Label: "6點", Assignees: point6, UserIDs: point6Users},
		{Label: "9點", Assignees: point9, UserIDs: point9Users},
		{Label: "12點", Assignees: point12, UserIDs: point12Users},
	}
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
func getBoss3Tasks(day DayAssignment) []string {
	totalCount := len(day.DPS)
	if day.Tank != "缺坦" {
		totalCount++
	}
	if day.Healer != "缺補" {
		totalCount++
	}

	tasks := []string{
		"狀態支援",
	}

	if totalCount >= 7 {
		tasks = append(tasks, "烙印")
		tasks = append(tasks, "煙")
	} else {
		tasks = append(tasks, "烙印、煙")
	}

	tasks = append(tasks,
		"鉤拳、貓蒼",
		"80%刻印、鯨魚",
		"40%刻印、黃道、支援箭",
	)

	return tasks
}

func assignBoss3(day DayAssignment, history map[string][]string) []WorkAssignment {
	assignments := []WorkAssignment{}
	remaining := append([]string{}, day.DPS...)

	tankTask := "缺坦"
	tankUserID := ""
	if day.Tank != "缺坦" {
		tankTask = getDisplayName(day.Tank)
		tankUserID = day.Tank
	}

	pickAnyUser := func(taskKey string) (string, string) {
		userID, rest := pickPreferredAny(taskKey, remaining, history)
		if userID == "" {
			return "", "缺人"
		}
		remaining = rest
		return userID, getDisplayName(userID)
	}

	pickNewbieFirst := func(taskKey string) (string, string) {
		userID, rest := pickPreferredUser(taskKey, remaining, history, isNewbie)
		if userID != "" {
			remaining = rest
			return userID, getDisplayName(userID)
		}
		return pickAnyUser(taskKey)
	}

	pickExperiencedFirst := func(taskKey string) (string, string) {
		userID, rest := pickPreferredUser(taskKey, remaining, history, isExperienced)
		if userID != "" {
			remaining = rest
			return userID, getDisplayName(userID)
		}
		return pickAnyUser(taskKey)
	}

	assignments = append(assignments, WorkAssignment{
		Label:    "60%坦克工作",
		Assignee: tankTask,
		UserID:   tankUserID,
	})

	for _, task := range getBoss3Tasks(day) {
		switch task {
		case "你原本程式裡那幾個新手優先工作":
			userID, name := pickNewbieFirst(task)
			assignments = append(assignments, WorkAssignment{Label: task, Assignee: name, UserID: userID})
		case "你原本程式裡那幾個老手優先工作":
			userID, name := pickExperiencedFirst(task)
			assignments = append(assignments, WorkAssignment{Label: task, Assignee: name, UserID: userID})
		default:
			userID, name := pickAnyUser(task)
			assignments = append(assignments, WorkAssignment{Label: task, Assignee: name, UserID: userID})
		}
	}

	return assignments
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

func buildWeekTaskAssignments(weekKey string) WeekTaskAssignments {
	weekAssignment := buildWeekAssignment(weekKey)

	result := WeekTaskAssignments{
		Days: map[string]DayTaskAssignments{},
	}

	history := map[string][]string{}

	for _, dayKey := range getOrderedDayKeys() {
		day := weekAssignment.Days[dayKey]

		boss1 := assignBoss1(day, history)
		recordWorkAssignments(history, boss1)

		boss2 := assignBoss2Group(day, history)
		for _, group := range boss2 {
			for i, userID := range group.UserIDs {
				recordHistory(history, group.Label+"#"+string(rune('0'+i)), userID)
			}
		}

		boss3 := assignBoss3(day, history)
		recordWorkAssignments(history, boss3)

		result.Days[dayKey] = DayTaskAssignments{
			Boss1: boss1,
			Boss2: boss2,
			Boss3: boss3,
		}
	}

	return result
}
