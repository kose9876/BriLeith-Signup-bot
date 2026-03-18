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

func setupBotHandlers(dg *discordgo.Session) {
	dg.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		fmt.Println("Bot 已登入：", s.State.User.Username)
	})

	dg.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.Bot {
			return
		}

		if strings.TrimSpace(m.Content) == "!ping" {
			if _, err := s.ChannelMessageSend(m.ChannelID, "Pong"); err != nil {
				fmt.Println("send ping reply failed:", err)
			}
		}
	})

	dg.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Type == discordgo.InteractionApplicationCommand && i.ApplicationCommandData().Name == "help" {
			handleHelpCommand(s, i)
			return
		}

		if handleAdminCommand(s, i) {
			return
		}

		if i.Type == discordgo.InteractionApplicationCommand {
			handleApplicationCommand(s, i)
			return
		}

		if i.Type == discordgo.InteractionMessageComponent {
			handleMessageComponent(s, i)
		}
	})
}

func handleApplicationCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.ApplicationCommandData().Name {
	case "signup":
		handleSignupCommand(s, i)
	case "setgamename":
		handleSetGameNameCommand(s, i)
	case "whatrole":
		handleWhatRoleCommand(s, i)
	case "setrole":
		handleSetRoleCommand(s, i)
	case "summary":
		handleSummaryCommand(s, i)
	}
}

func handleSignupCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	userID := i.Member.User.ID
	if !canUserManageSignup(userID) {
		respondEphemeral(s, i, "目前不可報名。")
		return
	}

	weekKey := getManagedSignupWeekKey()
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content:    buildSignupPanelContent(weekKey),
			Components: buildSignupComponents(),
		},
	})
	if err != nil {
		fmt.Println("signup command failed:", err)
	}
}

func handleSetGameNameCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	gameName := i.ApplicationCommandData().Options[0].StringValue()
	userID := i.Member.User.ID
	username := i.Member.User.Username
	updateGameName(userID, username, gameName)

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "已更新遊戲名稱：" + gameName,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		fmt.Println("setgamename command failed:", err)
	}
}

func handleWhatRoleCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	targetUser := i.ApplicationCommandData().Options[0].UserValue(s)
	profile, exists := userProfiles[targetUser.ID]

	content := "查無此玩家的職業設定。"
	if exists && profile.MainRole != "" {
		content = fmt.Sprintf(
			"%s\n主職：%s\n副職：%s\n群組：%s",
			profile.DisplayName,
			getRoleLabel(profile.MainRole),
			getRoleLabel(profile.SubRole),
			getGroupLabel(profile.Group),
		)
	}

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
		},
	})
	if err != nil {
		fmt.Println("whatrole command failed:", err)
	}
}

func handleSetRoleCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	userID := i.Member.User.ID
	username := i.Member.User.Username
	profile, exists := userProfiles[userID]
	if !exists || profile.GameName == "" {
		respondEphemeral(s, i, "請先使用 /setgamename 設定遊戲名稱。")
		return
	}

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content:    buildSetRoleContent(userID, username),
			Components: buildSetRoleComponents(),
			Flags:      discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		fmt.Println("setrole command failed:", err)
	}
}

func handleSummaryCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	weekKey := getCurrentWeekKey()
	content := getWeekRangeText(weekKey) + " 分配摘要\n\n請選擇要查看的日期。"

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content:    content,
			Components: buildSummaryComponents(),
		},
	})
	if err != nil {
		fmt.Println("summary command failed:", err)
	}
}

func handleMessageComponent(s *discordgo.Session, i *discordgo.InteractionCreate) {
	customID := i.MessageComponentData().CustomID

	switch {
	case strings.HasPrefix(customID, "role_"):
		handleRoleComponent(s, i, customID)
	case strings.HasPrefix(customID, "group_"):
		handleGroupComponent(s, i, customID)
	case strings.HasPrefix(customID, "cape_"):
		handleCapeComponent(s, i, customID)
	case strings.HasPrefix(customID, "summary_"):
		handleSummaryComponent(s, i, customID)
	case strings.HasPrefix(customID, "test_summary_"):
		handleTestSummaryComponent(s, i, customID)
	case strings.HasPrefix(customID, "test_day_"):
		handleTestSignupComponent(s, i, customID)
	case strings.HasPrefix(customID, "day_"):
		handleSignupComponent(s, i, customID)
	}
}

func handleRoleComponent(s *discordgo.Session, i *discordgo.InteractionCreate, customID string) {
	userID := i.Member.User.ID
	username := i.Member.User.Username

	parts := strings.Split(customID, "_")
	if len(parts) != 3 {
		return
	}

	err := updateUserRole(userID, username, parts[1], parts[2])
	content := buildSetRoleContent(userID, username)
	if err != nil {
		content += "\n\n錯誤：" + err.Error()
	}

	respErr := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Content:    content,
			Components: buildSetRoleComponents(),
		},
	})
	if respErr != nil {
		fmt.Println("update setrole message failed:", respErr)
	}
}

func handleGroupComponent(s *discordgo.Session, i *discordgo.InteractionCreate, customID string) {
	userID := i.Member.User.ID
	username := i.Member.User.Username
	updateUserGroup(userID, username, strings.TrimPrefix(customID, "group_"))

	respErr := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Content:    buildSetRoleContent(userID, username),
			Components: buildSetRoleComponents(),
		},
	})
	if respErr != nil {
		fmt.Println("update setrole group message failed:", respErr)
	}
}

func handleCapeComponent(s *discordgo.Session, i *discordgo.InteractionCreate, customID string) {
	userID := i.Member.User.ID
	username := i.Member.User.Username
	updateUserCape(userID, username, customID == "cape_yes")

	respErr := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Content:    buildSetRoleContent(userID, username),
			Components: buildSetRoleComponents(),
		},
	})
	if respErr != nil {
		fmt.Println("update setrole cape message failed:", respErr)
	}
}

func handleSummaryComponent(s *discordgo.Session, i *discordgo.InteractionCreate, customID string) {
	weekKey := getCurrentWeekKey()
	dayKey := strings.TrimPrefix(customID, "summary_")
	content := buildFullDaySummaryText(weekKey, dayKey)

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
		},
	})
	if err != nil {
		fmt.Println("summary detail failed:", err)
	}
}

func handleTestSummaryComponent(s *discordgo.Session, i *discordgo.InteractionCreate, customID string) {
	weekKey := getSignupWeekKeyAt(nowInBotLocation())
	dayKey := strings.TrimPrefix(customID, "test_summary_")
	content := buildFullDaySummaryTextFromStore(testWeeklySignups, weekKey, dayKey)

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
		},
	})
	if err != nil {
		fmt.Println("test summary detail failed:", err)
	}
}

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
	selectedDay := strings.TrimPrefix(customID, "test_")
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

func respondComponentEphemeral(s *discordgo.Session, i *discordgo.InteractionCreate, content string) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		fmt.Println("component response failed:", err)
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
		lines = append(lines, fmt.Sprintf("<@%s>：%s", userID, strings.Join(dayTexts, "、")))
	}

	return strings.Join(lines, "\n")
}
