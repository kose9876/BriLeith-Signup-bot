package main

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

func handleTestSignupComponent(s *discordgo.Session, i *discordgo.InteractionCreate, customID string) {
	userID := i.Member.User.ID
	profile, exists := userProfiles[userID]
	if !exists || profile.GameName == "" {
		respondComponentEphemeral(s, i, "請先設定遊戲名稱後再進行測試報名。")
		return
	}
	if profile.MainRole == "" {
		respondComponentEphemeral(s, i, "請先設定主職後再進行測試報名。")
		return
	}

	weekKey := getSignupWeekKeyAt(nowInBotLocation())
	selectedDay := customID[len("test_"):]
	if selectedDay == "day_all" {
		addAllAvailableDays(testWeeklySignups, saveTestSignups, weekKey, userID)
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Content:    buildTestSignupPanelContent(weekKey),
				Components: buildTestSignupComponents(),
			},
		})
		if err != nil {
			fmt.Println("test signup all update failed:", err)
		}
		return
	}

	if isSignupDayFull(testWeeklySignups, weekKey, userID, selectedDay) {
		respondComponentEphemeral(s, i, fmt.Sprintf("%s 已經額滿 %d 人，無法再報名。", dayLabels[selectedDay], maxSignupUsersPerDay))
		return
	}

	toggleTestUserSignup(weekKey, userID, selectedDay)
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Content:    buildTestSignupPanelContent(weekKey),
			Components: buildTestSignupComponents(),
		},
	})
	if err != nil {
		fmt.Println("test signup update failed:", err)
	}
}
