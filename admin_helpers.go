package main

import (
	"fmt"
	"sort"
	"strings"
	"unicode"

	"github.com/bwmarrin/discordgo"
)

func respondEphemeral(s *discordgo.Session, i *discordgo.InteractionCreate, content string) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		fmt.Println("回覆互動訊息失敗:", err)
	}
}

func requireAdmin(s *discordgo.Session, i *discordgo.InteractionCreate) bool {
	if i.Member == nil || i.Member.User == nil || !isAdminUser(i.Member.User.ID) {
		respondEphemeral(s, i, "你不是管理員，無法使用這個指令。")
		return false
	}

	return true
}

func addUserSignupDay(weekKey string, userID string, day string) {
	if weeklySignups[weekKey] == nil {
		weeklySignups[weekKey] = map[string][]string{}
	}

	for _, existingDay := range weeklySignups[weekKey][userID] {
		if existingDay == day {
			return
		}
	}

	weeklySignups[weekKey][userID] = append(weeklySignups[weekKey][userID], day)
	sortUserDays(weeklySignups[weekKey][userID])
	saveSignups()
}

func removeUserSignupDay(weekKey string, userID string, day string) bool {
	if weeklySignups[weekKey] == nil {
		return false
	}

	days := weeklySignups[weekKey][userID]
	for idx, existingDay := range days {
		if existingDay != day {
			continue
		}

		updated := append([]string{}, days[:idx]...)
		updated = append(updated, days[idx+1:]...)
		if len(updated) == 0 {
			delete(weeklySignups[weekKey], userID)
		} else {
			weeklySignups[weekKey][userID] = updated
		}
		saveSignups()
		return true
	}

	return false
}

func removeUserSignupWeek(weekKey string, userID string) {
	if weeklySignups[weekKey] == nil {
		return
	}

	delete(weeklySignups[weekKey], userID)
	saveSignups()
}

func formatSignupDays(days []string) string {
	if len(days) == 0 {
		return "尚未報名"
	}

	labels := make([]string, 0, len(days))
	for _, day := range days {
		labels = append(labels, getDayLabel(day))
	}

	return strings.Join(labels, "、")
}

func getSignupAccessText(userID string) string {
	if isSignupBlocked(userID) {
		return "已禁止自行報名"
	}

	return "可自行報名"
}

func emptyFallback(value string) string {
	if strings.TrimSpace(value) == "" {
		return "未設定"
	}

	return value
}

func boolText(value bool, yes string, no string) string {
	if value {
		return yes
	}

	return no
}

func getDayLabel(day string) string {
	if label, exists := dayLabels[day]; exists {
		return label
	}

	return day
}

func normalizeDiscordUserID(value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	trimmed = strings.TrimPrefix(trimmed, "<@")
	trimmed = strings.TrimPrefix(trimmed, "!")
	trimmed = strings.TrimSuffix(trimmed, ">")
	if trimmed == "" {
		return "", fmt.Errorf("請輸入有效的 Discord user ID 或 @mention。")
	}

	for _, r := range trimmed {
		if !unicode.IsDigit(r) {
			return "", fmt.Errorf("請輸入有效的 Discord user ID 或 @mention。")
		}
	}

	return trimmed, nil
}

func normalizePlayerLookupKey(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func resolveRegisteredPlayer(value string) (string, UserProfile, bool, error) {
	if userID, err := normalizeDiscordUserID(value); err == nil {
		profile, exists := userProfiles[userID]
		if exists {
			return userID, profile, true, nil
		}

		return userID, UserProfile{}, false, nil
	}

	lookup := normalizePlayerLookupKey(value)
	if lookup == "" {
		return "", UserProfile{}, false, fmt.Errorf("請輸入玩家 ID、@mention、遊戲名或顯示名。")
	}

	var exactMatches []playerMatch
	var partialMatches []playerMatch

	for userID, profile := range userProfiles {
		gameName := normalizePlayerLookupKey(profile.GameName)
		displayName := normalizePlayerLookupKey(profile.DisplayName)

		if lookup == gameName || lookup == displayName {
			exactMatches = append(exactMatches, playerMatch{userID: userID, profile: profile})
			continue
		}

		if (gameName != "" && strings.Contains(gameName, lookup)) || (displayName != "" && strings.Contains(displayName, lookup)) {
			partialMatches = append(partialMatches, playerMatch{userID: userID, profile: profile})
		}
	}

	if len(exactMatches) == 1 {
		return exactMatches[0].userID, exactMatches[0].profile, true, nil
	}
	if len(exactMatches) > 1 {
		return "", UserProfile{}, false, fmt.Errorf("找到多位同名玩家，請改用 user ID。候選: %s", joinPlayerLabels(exactMatches))
	}
	if len(partialMatches) == 1 {
		return partialMatches[0].userID, partialMatches[0].profile, true, nil
	}
	if len(partialMatches) > 1 {
		return "", UserProfile{}, false, fmt.Errorf("找到多位相符玩家，請輸入更完整名稱或 user ID。候選: %s", joinPlayerLabels(partialMatches))
	}

	return "", UserProfile{}, false, fmt.Errorf("找不到玩家: %s", strings.TrimSpace(value))
}

func mentionUser(userID string) string {
	return fmt.Sprintf("<@%s>", userID)
}

func collectEnabledUserMentions(items map[string]bool) []string {
	userIDs := make([]string, 0, len(items))
	for userID, enabled := range items {
		if enabled {
			userIDs = append(userIDs, userID)
		}
	}

	sort.Strings(userIDs)

	result := make([]string, 0, len(userIDs))
	for _, userID := range userIDs {
		result = append(result, formatPlayerLabel(userID))
	}

	return result
}

func formatPlayerLabel(userID string) string {
	profile, exists := userProfiles[userID]
	if !exists {
		return userID
	}

	gameName := strings.TrimSpace(profile.GameName)
	displayName := strings.TrimSpace(profile.DisplayName)

	switch {
	case gameName != "" && displayName != "":
		return fmt.Sprintf("%s (%s / %s)", gameName, displayName, userID)
	case gameName != "":
		return fmt.Sprintf("%s (%s)", gameName, userID)
	case displayName != "":
		return fmt.Sprintf("%s (%s)", displayName, userID)
	default:
		return userID
	}
}

func joinPlayerLabels(matches []playerMatch) string {
	labels := make([]string, 0, len(matches))
	for _, item := range matches {
		labels = append(labels, formatPlayerLabel(item.userID))
	}
	return strings.Join(labels, "、")
}

func joinOrNone(items []string) string {
	if len(items) == 0 {
		return "無"
	}

	return strings.Join(items, "、")
}
