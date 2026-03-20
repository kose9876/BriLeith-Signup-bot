package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type AdminState struct {
	AdminUsers         map[string]bool `json:"admin_users"`
	TesterUsers        map[string]bool `json:"tester_users"`
	BlockedSignupUsers map[string]bool `json:"blocked_signup_users"`
}

type playerMatch struct {
	userID  string
	profile UserProfile
}

var adminState = AdminState{
	AdminUsers:         map[string]bool{},
	TesterUsers:        map[string]bool{},
	BlockedSignupUsers: map[string]bool{},
}

func loadAdminState() {
	if err := readJSONFile("admin_state.json", &adminState); err != nil {
		fmt.Println("讀取 admin_state.json 失敗:", err)
		adminState = AdminState{
			AdminUsers:         map[string]bool{},
			TesterUsers:        map[string]bool{},
			BlockedSignupUsers: map[string]bool{},
		}
		return
	}

	if adminState.AdminUsers == nil {
		adminState.AdminUsers = map[string]bool{}
	}
	if adminState.TesterUsers == nil {
		adminState.TesterUsers = map[string]bool{}
	}
	if adminState.BlockedSignupUsers == nil {
		adminState.BlockedSignupUsers = map[string]bool{}
	}
}

func saveAdminState() {
	data, err := json.MarshalIndent(adminState, "", "  ")
	if err != nil {
		fmt.Println("轉換 admin_state.json 失敗:", err)
		return
	}

	err = os.WriteFile("admin_state.json", data, 0644)
	if err != nil {
		fmt.Println("寫入 admin_state.json 失敗:", err)
	}
}

func isAdminUser(userID string) bool {
	for _, adminUserID := range appConfig.AdminUserIDs {
		if adminUserID == userID {
			return true
		}
	}

	return adminState.AdminUsers[userID]
}

func isSignupBlocked(userID string) bool {
	return adminState.BlockedSignupUsers[userID]
}

func isTesterUser(userID string) bool {
	if isAdminUser(userID) {
		return true
	}
	return adminState.TesterUsers[userID]
}
