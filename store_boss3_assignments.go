package main

import "fmt"

const boss3AssignmentsFile = "boss3_assignments.json"

var boss3Assignments = map[string]map[string]map[string]string{}

func loadBoss3Assignments() {
	if err := readJSONFile(boss3AssignmentsFile, &boss3Assignments); err != nil {
		fmt.Println("load boss3_assignments.json failed:", err)
		boss3Assignments = map[string]map[string]map[string]string{}
	}
}

func saveBoss3Assignments() {
	if err := writeJSONFile(boss3AssignmentsFile, boss3Assignments); err != nil {
		fmt.Println("write boss3_assignments.json failed:", err)
	}
}

func setBoss3Assignment(weekKey string, dayKey string, taskLabel string, userID string, mode Boss3OverrideMode) {
	if boss3Assignments[weekKey] == nil {
		boss3Assignments[weekKey] = map[string]map[string]string{}
	}
	if boss3Assignments[weekKey][dayKey] == nil {
		boss3Assignments[weekKey][dayKey] = map[string]string{}
	}

	boss3Assignments[weekKey][dayKey][taskLabel] = encodeBoss3Override(mode, userID)
	saveBoss3Assignments()
}

func clearBoss3Assignment(weekKey string, dayKey string, taskLabel string) bool {
	weekAssignments := boss3Assignments[weekKey]
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
		delete(boss3Assignments, weekKey)
	}

	saveBoss3Assignments()
	return true
}

func buildBoss3AssignmentsFromStore(signups map[string]map[string][]string, overrides map[string]map[string]map[string]string, save func(), isAssigned func(string, string, string) bool, weekKey string, dayKey string) []WorkAssignment {
	assignment := buildWeekAssignmentFromStore(signups, weekKey)
	day := assignment.Days[dayKey]
	assignments := assignBoss3(day, map[string][]string{})

	weekOverrides := overrides[weekKey]
	if weekOverrides == nil {
		return assignments
	}

	return applyBoss3AssignmentOverridesWithStore(
		overrides,
		save,
		func(userID string) bool {
			return isAssigned(weekKey, dayKey, userID)
		},
		weekKey,
		dayKey,
		assignments,
		weekOverrides[dayKey],
	)
}

func buildFormalBoss3Assignments(weekKey string, dayKey string) []WorkAssignment {
	return buildBoss3AssignmentsFromStore(weeklySignups, boss3Assignments, saveBoss3Assignments, userAssignedToFormalDay, weekKey, dayKey)
}

func userAssignedToFormalDay(weekKey string, dayKey string, userID string) bool {
	return userHasSignupDay(weeklySignups, weekKey, userID, dayKey)
}
