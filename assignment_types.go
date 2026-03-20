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
