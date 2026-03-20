package main

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

var weeklySignups = map[string]map[string][]string{}

var dayLabels = map[string]string{
	"day_mon": "周一",
	"day_tue": "周二",
	"day_wed": "周三",
	"day_thu": "周四",
	"day_fri": "周五",
	"day_sat": "周六",
	"day_sun": "周日",
}

var dayOrder = map[string]int{
	"day_mon": 1,
	"day_tue": 2,
	"day_wed": 3,
	"day_thu": 4,
	"day_fri": 5,
	"day_sat": 6,
	"day_sun": 7,
}

func handleSignupComponent(s *discordgo.Session, i *discordgo.InteractionCreate, selectedDay string) {
	userID := i.Member.User.ID
	if !canUserManageSignup(userID) {
		respondComponentEphemeral(s, i, "目前不可報名。")
		return
	}
	if isSignupBlocked(userID) {
		respondComponentEphemeral(s, i, "你目前不能自行報名，若要補登請聯絡管理員。")
		return
	}

	profile, exists := userProfiles[userID]
	if !exists || profile.GameName == "" {
		respondComponentEphemeral(s, i, "請先使用 /setgamename 設定遊戲名稱。")
		return
	}
	if profile.MainRole == "" {
		respondComponentEphemeral(s, i, "請先使用 /setrole 設定主職。")
		return
	}

	weekKey := getManagedSignupWeekKey()
	if selectedDay == "day_all" {
		addAllAvailableDays(weeklySignups, saveSignups, weekKey, userID)
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Content:    buildSignupPanelContent(weekKey),
				Components: i.Message.Components,
			},
		})
		if err != nil {
			fmt.Println("signup all update failed:", err)
		}
		return
	}

	if isSignupDayFull(weeklySignups, weekKey, userID, selectedDay) {
		respondComponentEphemeral(s, i, fmt.Sprintf("%s 已經額滿 %d 人，無法再報名。", dayLabels[selectedDay], maxSignupUsersPerDay))
		return
	}

	toggleUserSignup(weekKey, userID, selectedDay)
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Content:    buildSignupPanelContent(weekKey),
			Components: i.Message.Components,
		},
	})
	if err != nil {
		fmt.Println("signup update failed:", err)
	}
}

func toggleUserSignup(weekKey string, userID string, day string) {
	if weeklySignups[weekKey] == nil {
		weeklySignups[weekKey] = map[string][]string{}
	}

	currentDays := weeklySignups[weekKey][userID]
	for i, existingDay := range currentDays {
		if existingDay != day {
			continue
		}
		weeklySignups[weekKey][userID] = append(currentDays[:i], currentDays[i+1:]...)
		sortUserDays(weeklySignups[weekKey][userID])
		saveSignups()
		return
	}

	weeklySignups[weekKey][userID] = append(weeklySignups[weekKey][userID], day)
	sortUserDays(weeklySignups[weekKey][userID])
	saveSignups()
}

func addAllAvailableDays(signups map[string]map[string][]string, save func(), weekKey string, userID string) {
	if signups[weekKey] == nil {
		signups[weekKey] = map[string][]string{}
	}

	currentDays := append([]string{}, signups[weekKey][userID]...)
	seen := map[string]bool{}
	for _, day := range currentDays {
		seen[day] = true
	}

	for _, day := range getOrderedDayKeys() {
		if seen[day] {
			continue
		}
		if countSignupUsersForDay(signups, weekKey, day) >= maxSignupUsersPerDay {
			continue
		}
		currentDays = append(currentDays, day)
	}

	sortUserDays(currentDays)
	signups[weekKey][userID] = currentDays
	save()
}

func buildSignupSummary(weekKey string) string {
	return buildSignupSummaryFromStore(weeklySignups, weekKey, "目前還沒有人報名。")
}

func buildSignupSummaryFromStore(signups map[string]map[string][]string, weekKey string, emptyMessage string) string {
	users := signups[weekKey]
	if len(users) == 0 {
		return emptyMessage
	}

	lines := make([]string, 0, len(users))
	for userID, days := range users {
		dayTexts := make([]string, 0, len(days))
		for _, day := range days {
			dayTexts = append(dayTexts, dayLabels[day])
		}
		lines = append(lines, fmt.Sprintf("%s：%s", getDisplayName(userID), strings.Join(dayTexts, "、")))
	}

	return strings.Join(lines, "\n")
}
