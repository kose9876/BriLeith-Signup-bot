package main

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func handleAdminBoss3AssignCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
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
	if !userAssignedToFormalDay(weekKey, dayKey, userID) {
		respondEphemeral(s, i, fmt.Sprintf("%s 沒有報名這一天的正式名單。", formatPlayerLabel(userID)))
		return
	}

	currentAssignments := buildFormalBoss3Assignments(weekKey, dayKey)
	if indexOfBoss3Assignment(currentAssignments, taskLabel) == -1 {
		respondEphemeral(s, i, "這個工作目前不在該日的三王分配內，請先看正式摘要確認工作名稱。")
		return
	}
	if mode != Boss3OverrideSwap && mode != Boss3OverrideAdd {
		respondEphemeral(s, i, "未知的 mode 參數。")
		return
	}

	setBoss3Assignment(weekKey, dayKey, taskLabel, userID, mode)
	updatedAssignments := buildFormalBoss3Assignments(weekKey, dayKey)

	respondEphemeral(s, i, fmt.Sprintf(
		"已更新正式版三王工作。\n日期: %s\n工作: %s\n模式: %s\n玩家: %s\n\n目前三王分配:\n%s",
		getDayLabel(dayKey),
		taskLabel,
		getBoss3OverrideModeLabel(mode),
		formatPlayerLabel(userID),
		buildBoss3TaskTextFromAssignments(updatedAssignments),
	))
}

func handleAdminBoss3ClearCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
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

	if !clearBoss3Assignment(weekKey, dayKey, taskLabel) {
		respondEphemeral(s, i, "這個工作目前沒有正式版手動覆寫紀錄。")
		return
	}

	updatedAssignments := buildFormalBoss3Assignments(weekKey, dayKey)
	respondEphemeral(s, i, fmt.Sprintf(
		"已清除正式版三王工作覆寫。\n日期: %s\n工作: %s\n\n目前三王分配:\n%s",
		getDayLabel(dayKey),
		taskLabel,
		buildBoss3TaskTextFromAssignments(updatedAssignments),
	))
}
