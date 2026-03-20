package main

type WorkAssignment struct {
	Label          string
	Assignee       string
	UserID         string
	ExtraAssignees []string
	ExtraUserIDs   []string
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

func fillToTwo(assignees []string) []string {
	for len(assignees) < 2 {
		assignees = append(assignees, "缺人")
	}
	return assignees
}
