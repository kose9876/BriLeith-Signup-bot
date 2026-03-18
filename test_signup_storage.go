package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/bwmarrin/discordgo"
)

const testSignupsFile = "test_signups.json"

var testWeeklySignups = map[string]map[string][]string{}

func loadTestSignups() {
	if err := readJSONFile(testSignupsFile, &testWeeklySignups); err != nil {
		fmt.Println("load test_signups.json failed:", err)
		testWeeklySignups = map[string]map[string][]string{}
		return
	}

	for weekKey, users := range testWeeklySignups {
		for userID, days := range users {
			sortUserDays(days)
			testWeeklySignups[weekKey][userID] = days
		}
	}
}

func saveTestSignups() {
	data, err := json.MarshalIndent(testWeeklySignups, "", "  ")
	if err != nil {
		fmt.Println("marshal test_signups.json failed:", err)
		return
	}

	if err := os.WriteFile(testSignupsFile, data, 0644); err != nil {
		fmt.Println("write test_signups.json failed:", err)
	}
}

func buildTestSignupComponents() []discordgo.MessageComponent {
	return []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{Label: dayLabels["day_mon"], Style: discordgo.PrimaryButton, CustomID: "test_day_mon"},
				discordgo.Button{Label: dayLabels["day_tue"], Style: discordgo.PrimaryButton, CustomID: "test_day_tue"},
				discordgo.Button{Label: dayLabels["day_wed"], Style: discordgo.PrimaryButton, CustomID: "test_day_wed"},
				discordgo.Button{Label: dayLabels["day_thu"], Style: discordgo.PrimaryButton, CustomID: "test_day_thu"},
				discordgo.Button{Label: dayLabels["day_fri"], Style: discordgo.PrimaryButton, CustomID: "test_day_fri"},
			},
		},
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{Label: dayLabels["day_sat"], Style: discordgo.SuccessButton, CustomID: "test_day_sat"},
				discordgo.Button{Label: dayLabels["day_sun"], Style: discordgo.SuccessButton, CustomID: "test_day_sun"},
			},
		},
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{Label: "我要打全部", Style: discordgo.DangerButton, CustomID: "test_day_all"},
			},
		},
	}
}

func buildTestSignupPanelContent(weekKey string) string {
	return fmt.Sprintf(
		"%s 報名\n\n每一天最多 %d 人，額滿後停止該日報名。\n請選擇你要報名的日期：\n\n%s",
		getWeekRangeText(weekKey),
		maxSignupUsersPerDay,
		buildTestSignupSummary(weekKey),
	)
}

func toggleTestUserSignup(weekKey string, userID string, day string) {
	if testWeeklySignups[weekKey] == nil {
		testWeeklySignups[weekKey] = map[string][]string{}
	}

	currentDays := testWeeklySignups[weekKey][userID]
	for i, existingDay := range currentDays {
		if existingDay != day {
			continue
		}

		testWeeklySignups[weekKey][userID] = append(currentDays[:i], currentDays[i+1:]...)
		sortUserDays(testWeeklySignups[weekKey][userID])
		saveTestSignups()
		return
	}

	testWeeklySignups[weekKey][userID] = append(testWeeklySignups[weekKey][userID], day)
	sortUserDays(testWeeklySignups[weekKey][userID])
	saveTestSignups()
}

func buildTestSignupSummary(weekKey string) string {
	return buildSignupSummaryFromStore(testWeeklySignups, weekKey, "目前還沒有人報名。")
}

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

	userOption := findOption(options, "user")
	dayOption := findOption(options, "day")
	if userOption == nil || dayOption == nil {
		respondEphemeral(s, i, "缺少必要參數。")
		return
	}

	user := userOption.UserValue(s)
	day := dayOption.StringValue()

	addTestUserSignupDay(getManagedSignupWeekKey(), user.ID, day)

	respondEphemeral(s, i, fmt.Sprintf(
		"已幫 <@%s> 手動加入測試報名 %s。",
		user.ID,
		getDayLabel(day),
	))
}

func handleAdminTestUnsignupCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options

	userOption := findOption(options, "user")
	dayOption := findOption(options, "day")
	if userOption == nil || dayOption == nil {
		respondEphemeral(s, i, "缺少必要參數。")
		return
	}

	user := userOption.UserValue(s)
	day := dayOption.StringValue()

	if !removeTestUserSignupDay(getManagedSignupWeekKey(), user.ID, day) {
		respondEphemeral(s, i, fmt.Sprintf(
			"<@%s> 本週測試報名沒有 %s。",
			user.ID,
			getDayLabel(day),
		))
		return
	}

	respondEphemeral(s, i, fmt.Sprintf(
		"已幫 <@%s> 取消測試報名 %s。",
		user.ID,
		getDayLabel(day),
	))
}

func addTestUserSignupDay(weekKey string, userID string, day string) {
	if testWeeklySignups[weekKey] == nil {
		testWeeklySignups[weekKey] = map[string][]string{}
	}

	for _, existingDay := range testWeeklySignups[weekKey][userID] {
		if existingDay == day {
			return
		}
	}

	testWeeklySignups[weekKey][userID] = append(testWeeklySignups[weekKey][userID], day)
	sortUserDays(testWeeklySignups[weekKey][userID])
	saveTestSignups()
}

func removeTestUserSignupDay(weekKey string, userID string, day string) bool {
	if testWeeklySignups[weekKey] == nil {
		return false
	}

	days := testWeeklySignups[weekKey][userID]
	for idx, existingDay := range days {
		if existingDay != day {
			continue
		}

		testWeeklySignups[weekKey][userID] = append(days[:idx], days[idx+1:]...)
		if len(testWeeklySignups[weekKey][userID]) == 0 {
			delete(testWeeklySignups[weekKey], userID)
		}
		saveTestSignups()
		return true
	}

	return false
}
