package main

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func handleAdminTestSignupPanelCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	weekKey := getSignupWeekKeyAt(nowInBotLocation())
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content:    buildTestSignupPanelContent(weekKey),
			Components: buildTestSignupComponents(),
			Flags:      discordgo.MessageFlagsEphemeral,
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
			Flags:      discordgo.MessageFlagsEphemeral,
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
		respondEphemeral(s, i, "這個玩家還沒有資料，請先使用 /a_addplayer 建立名單。")
		return
	}
	day := dayOption.StringValue()
	weekKey := getManagedSignupWeekKey()

	if isSignupDayFull(testWeeklySignups, weekKey, userID, day) {
		respondEphemeral(s, i, fmt.Sprintf(
			"%s 測試報名已額滿 %d 人，無法再手動加入。",
			getDayLabel(day),
			maxSignupUsersPerDay,
		))
		return
	}

	addTestUserSignupDay(weekKey, userID, day)

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
		respondEphemeral(s, i, "這個玩家還沒有資料，請先使用 /a_addplayer 建立名單。")
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

func handleAdminTestBoss3AssignCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options
	dayOption := findOption(options, "day")
	taskOption := findOption(options, "task")
	modeOption := findOption(options, "mode")
	playerOption := findOption(options, "player")
	if dayOption == nil || taskOption == nil || modeOption == nil || playerOption == nil {
		respondEphemeral(s, i, "缺少必要參數。")
		return
	}

	weekKey := getManagedSignupWeekKey()
	dayKey := dayOption.StringValue()
	taskLabel := strings.TrimSpace(taskOption.StringValue())
	mode := Boss3OverrideMode(strings.TrimSpace(modeOption.StringValue()))
	userID, _, exists, err := resolveRegisteredPlayer(playerOption.StringValue())
	if err != nil {
		respondEphemeral(s, i, err.Error())
		return
	}
	if !exists {
		respondEphemeral(s, i, "這個玩家還沒有資料，請先使用 /a_addplayer 建立名單。")
		return
	}
	if !userAssignedToTestDay(weekKey, dayKey, userID) {
		respondEphemeral(s, i, fmt.Sprintf("%s 沒有報名這一天的測試名單。", formatPlayerLabel(userID)))
		return
	}

	currentAssignments := buildTestBoss3Assignments(weekKey, dayKey)
	if indexOfBoss3Assignment(currentAssignments, taskLabel) == -1 {
		respondEphemeral(s, i, "這個工作目前不在該日的三王分配內，請先看測試摘要確認工作名稱。")
		return
	}

	if mode != Boss3OverrideSwap && mode != Boss3OverrideAdd {
		respondEphemeral(s, i, "未知的 mode 參數。")
		return
	}

	setTestBoss3Assignment(weekKey, dayKey, taskLabel, userID, mode)
	updatedAssignments := buildTestBoss3Assignments(weekKey, dayKey)

	respondEphemeral(s, i, fmt.Sprintf(
		"已更新測試版三王工作。\n日期: %s\n工作: %s\n模式: %s\n玩家: %s\n\n目前三王分配:\n%s",
		getDayLabel(dayKey),
		taskLabel,
		getBoss3OverrideModeLabel(mode),
		formatPlayerLabel(userID),
		buildBoss3TaskTextFromAssignments(updatedAssignments),
	))
}

func handleAdminTestBoss3ClearCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options
	dayOption := findOption(options, "day")
	taskOption := findOption(options, "task")
	if dayOption == nil || taskOption == nil {
		respondEphemeral(s, i, "缺少必要參數。")
		return
	}

	weekKey := getManagedSignupWeekKey()
	dayKey := dayOption.StringValue()
	taskLabel := strings.TrimSpace(taskOption.StringValue())

	if !clearTestBoss3Assignment(weekKey, dayKey, taskLabel) {
		respondEphemeral(s, i, "這個工作目前沒有手動覆寫紀錄。")
		return
	}

	updatedAssignments := buildTestBoss3Assignments(weekKey, dayKey)
	respondEphemeral(s, i, fmt.Sprintf(
		"已清除測試版三王工作覆寫。\n日期: %s\n工作: %s\n\n目前三王分配:\n%s",
		getDayLabel(dayKey),
		taskLabel,
		buildBoss3TaskTextFromAssignments(updatedAssignments),
	))
}

func getBoss3OverrideModeLabel(mode Boss3OverrideMode) string {
	switch mode {
	case Boss3OverrideAdd:
		return "追加兼任"
	default:
		return "換位"
	}
}
