package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
)

func sortUserDays(days []string) {
	sort.Slice(days, func(i, j int) bool {
		return dayOrder[days[i]] < dayOrder[days[j]]
	})
}

func loadSignups() {
	if err := readJSONFile("signups.json", &weeklySignups); err != nil {
		fmt.Println("讀取 signups.json 失敗，將使用空資料:", err)
		weeklySignups = map[string]map[string][]string{}
		return
	}

	for weekKey, users := range weeklySignups {
		for userID, days := range users {
			sortUserDays(days)
			weeklySignups[weekKey][userID] = days
		}
	}
}

func saveSignups() {
	data, err := json.MarshalIndent(weeklySignups, "", "  ")
	if err != nil {
		fmt.Println("轉換 signups JSON 失敗:", err)
		return
	}

	err = os.WriteFile("signups.json", data, 0644)
	if err != nil {
		fmt.Println("寫入 signups.json 失敗:", err)
	}
}
