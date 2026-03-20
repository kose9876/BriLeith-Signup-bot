package main

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

func handleAdminTestSignupPanelCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	weekKey := getSignupWeekKeyAt(nowInBotLocation())
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content:    buildTestSignupPanelContent(weekKey),
			Components: buildTestSignupComponents(),
		},
	})
	if err != nil {
		fmt.Println("admin test signup panel failed:", err)
	}
}

func handleAdminTestSummaryCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	weekKey := getSignupWeekKeyAt(nowInBotLocation())
	content := getWeekRangeText(weekKey) + " 分配摘要\n\n請選擇要查看的日期。"

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content:    content,
			Components: buildTestSummaryComponents(),
		},
	})
	if err != nil {
		fmt.Println("admin test summary failed:", err)
	}
}

func handleAdminTestSignupCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options

	playerOption := findOption(options, "player")
	dayOption := findOption(options, "day")
	if playerOption == nil || dayOption == nil {
		respondEphemeral(s, i, "缺少必要參數。")
		return
	}

	userID, _, exists, err := resolveRegisteredPlayer(playerOption.StringValue())
	if err != nil {
		respondEphemeral(s, i, err.Error())
		return
	}
	if !exists {
		respondEphemeral(s, i, "這個玩家還沒有資料，請先使用 /admin_addplayer 建立名單。")
		return
	}
	day := dayOption.StringValue()

	addTestUserSignupDay(getManagedSignupWeekKey(), userID, day)

	respondEphemeral(s, i, fmt.Sprintf(
		"已幫 %s 手動加入測試報名 %s。",
		formatPlayerLabel(userID),
		getDayLabel(day),
	))
}

func handleAdminTestUnsignupCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options

	playerOption := findOption(options, "player")
	dayOption := findOption(options, "day")
	if playerOption == nil || dayOption == nil {
		respondEphemeral(s, i, "缺少必要參數。")
		return
	}

	userID, _, exists, err := resolveRegisteredPlayer(playerOption.StringValue())
	if err != nil {
		respondEphemeral(s, i, err.Error())
		return
	}
	if !exists {
		respondEphemeral(s, i, "這個玩家還沒有資料，請先使用 /admin_addplayer 建立名單。")
		return
	}
	day := dayOption.StringValue()

	if !removeTestUserSignupDay(getManagedSignupWeekKey(), userID, day) {
		respondEphemeral(s, i, fmt.Sprintf(
			"%s 本週測試報名沒有 %s。",
			formatPlayerLabel(userID),
			getDayLabel(day),
		))
		return
	}

	respondEphemeral(s, i, fmt.Sprintf(
		"已幫 %s 取消測試報名 %s。",
		formatPlayerLabel(userID),
		getDayLabel(day),
	))
}
