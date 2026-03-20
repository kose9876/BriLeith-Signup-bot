package main

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

const testBoss3AssignmentsFile = "test_boss3_assignments.json"

var testBoss3Assignments = map[string]map[string]map[string]string{}

func loadTestBoss3Assignments() {
	if err := readJSONFile(testBoss3AssignmentsFile, &testBoss3Assignments); err != nil {
		fmt.Println("load test_boss3_assignments.json failed:", err)
		testBoss3Assignments = map[string]map[string]map[string]string{}
	}
}

func saveTestBoss3Assignments() {
	if err := writeJSONFile(testBoss3AssignmentsFile, testBoss3Assignments); err != nil {
		fmt.Println("write test_boss3_assignments.json failed:", err)
	}
}

func setTestBoss3Assignment(weekKey string, dayKey string, taskLabel string, userID string, mode Boss3OverrideMode) {
	if testBoss3Assignments[weekKey] == nil {
		testBoss3Assignments[weekKey] = map[string]map[string]string{}
	}
	if testBoss3Assignments[weekKey][dayKey] == nil {
		testBoss3Assignments[weekKey][dayKey] = map[string]string{}
	}

	testBoss3Assignments[weekKey][dayKey][taskLabel] = encodeBoss3Override(mode, userID)
	saveTestBoss3Assignments()
}

func clearTestBoss3Assignment(weekKey string, dayKey string, taskLabel string) bool {
	weekAssignments := testBoss3Assignments[weekKey]
	if weekAssignments == nil || weekAssignments[dayKey] == nil {
		return false
	}
	if _, exists := weekAssignments[dayKey][taskLabel]; !exists {
		return false
	}

	delete(weekAssignments[dayKey], taskLabel)
	if len(weekAssignments[dayKey]) == 0 {
		delete(weekAssignments, dayKey)
	}
	if len(weekAssignments) == 0 {
		delete(testBoss3Assignments, weekKey)
	}

	saveTestBoss3Assignments()
	return true
}

func buildTestBoss3Assignments(weekKey string, dayKey string) []WorkAssignment {
	assignment := buildWeekAssignmentFromStore(testWeeklySignups, weekKey)
	day := assignment.Days[dayKey]
	assignments := assignBoss3(day, map[string][]string{})

	overrides := testBoss3Assignments[weekKey]
	if overrides == nil {
		return assignments
	}

	return applyBoss3AssignmentOverrides(weekKey, dayKey, assignments, overrides[dayKey])
}

func applyBoss3AssignmentOverrides(weekKey string, dayKey string, assignments []WorkAssignment, overrides map[string]string) []WorkAssignment {
	if len(overrides) == 0 {
		return assignments
	}

	result := append([]WorkAssignment{}, assignments...)
	dirty := false
	for taskLabel, rawValue := range overrides {
		mode, userID := decodeBoss3Override(rawValue)
		taskIndex := indexOfBoss3Assignment(result, taskLabel)
		if taskIndex == -1 {
			delete(testBoss3Assignments[weekKey][dayKey], taskLabel)
			dirty = true
			continue
		}
		if userID != "" && !userAssignedToTestDay(weekKey, dayKey, userID) {
			delete(testBoss3Assignments[weekKey][dayKey], taskLabel)
			dirty = true
			continue
		}

		if mode == Boss3OverrideAdd {
			applyBoss3AddOverride(&result[taskIndex], userID)
			continue
		}

		previousUserID := result[taskIndex].UserID
		previousAssignee := result[taskIndex].Assignee
		sourceIndex := -1
		for idx := range result {
			if idx == taskIndex || result[idx].UserID != userID || userID == "" {
				continue
			}
			sourceIndex = idx
			break
		}

		result[taskIndex].UserID = userID
		if userID == "" {
			result[taskIndex].Assignee = "缺人"
			continue
		}
		result[taskIndex].Assignee = getDisplayName(userID)

		if sourceIndex != -1 {
			if previousUserID == "" {
				result[sourceIndex].UserID = ""
				result[sourceIndex].Assignee = "缺人"
			} else {
				result[sourceIndex].UserID = previousUserID
				result[sourceIndex].Assignee = previousAssignee
			}
		}
	}

	if dirty {
		saveTestBoss3Assignments()
	}

	return result
}

func indexOfBoss3Assignment(assignments []WorkAssignment, taskLabel string) int {
	for idx, assignment := range assignments {
		if assignment.Label == taskLabel {
			return idx
		}
	}
	return -1
}

func applyBoss3AddOverride(assignment *WorkAssignment, userID string) {
	if userID == "" {
		assignment.UserID = ""
		assignment.Assignee = "缺人"
		assignment.ExtraUserIDs = nil
		assignment.ExtraAssignees = nil
		return
	}

	if assignment.UserID == "" {
		assignment.UserID = userID
		assignment.Assignee = getDisplayName(userID)
		return
	}
	if assignment.UserID == userID || containsUser(assignment.ExtraUserIDs, userID) {
		return
	}

	assignment.ExtraUserIDs = append(assignment.ExtraUserIDs, userID)
	assignment.ExtraAssignees = append(assignment.ExtraAssignees, getDisplayName(userID))
}

func encodeBoss3Override(mode Boss3OverrideMode, userID string) string {
	if mode == "" {
		mode = Boss3OverrideSwap
	}
	return string(mode) + ":" + userID
}

func decodeBoss3Override(value string) (Boss3OverrideMode, string) {
	if strings.HasPrefix(value, string(Boss3OverrideAdd)+":") {
		return Boss3OverrideAdd, strings.TrimPrefix(value, string(Boss3OverrideAdd)+":")
	}
	if strings.HasPrefix(value, string(Boss3OverrideSwap)+":") {
		return Boss3OverrideSwap, strings.TrimPrefix(value, string(Boss3OverrideSwap)+":")
	}
	return Boss3OverrideSwap, value
}

func userAssignedToTestDay(weekKey string, dayKey string, userID string) bool {
	return userHasSignupDay(testWeeklySignups, weekKey, userID, dayKey)
}

func buildBoss3TaskChoices() []*discordgo.ApplicationCommandOptionChoice {
	return []*discordgo.ApplicationCommandOptionChoice{
		{Name: "狀態支援", Value: "狀態支援"},
		{Name: "烙印", Value: "烙印"},
		{Name: "煙", Value: "煙"},
		{Name: "鉤拳、貓蒼", Value: "鉤拳、貓蒼"},
		{Name: "80%刻印、鯨魚", Value: "80%刻印、鯨魚"},
		{Name: "40%刻印、黃道、支援箭", Value: "40%刻印、黃道、支援箭"},
	}
}

func buildBoss3OverrideModeChoices() []*discordgo.ApplicationCommandOptionChoice {
	return []*discordgo.ApplicationCommandOptionChoice{
		{Name: "換位", Value: string(Boss3OverrideSwap)},
		{Name: "追加兼任", Value: string(Boss3OverrideAdd)},
	}
}
