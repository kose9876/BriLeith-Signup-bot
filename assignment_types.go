package main

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

type Boss3OverrideMode string

const (
	Boss3OverrideSwap Boss3OverrideMode = "swap"
	Boss3OverrideAdd  Boss3OverrideMode = "add"
)
