package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/bwmarrin/discordgo"
)

type SignupState struct {
	ForcedPublicSignupOpen *bool `json:"forced_public_signup_open,omitempty"`
}

var signupState = SignupState{}

func loadSignupScheduleState() {
	if err := readJSONFile("signup_schedule_state.json", &signupState); err != nil {
		fmt.Println("load signup_schedule_state.json failed:", err)
		signupState = SignupState{}
		return
	}
}

func saveSignupScheduleState() {
	data, err := json.MarshalIndent(signupState, "", "  ")
	if err != nil {
		fmt.Println("marshal signup_schedule_state.json failed:", err)
		return
	}

	if err := os.WriteFile("signup_schedule_state.json", data, 0644); err != nil {
		fmt.Println("write signup_schedule_state.json failed:", err)
	}
}

func getBotLocation() *time.Location {
	loc, err := time.LoadLocation("Asia/Taipei")
	if err != nil {
		return time.Local
	}

	return loc
}

func nowInBotLocation() time.Time {
	return time.Now().In(getBotLocation())
}

func getCurrentWeekKeyAt(now time.Time) string {
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7
	}

	monday := now.AddDate(0, 0, -(weekday - 1))
	return monday.Format("2006-01-02")
}

func getSignupWeekKeyAt(now time.Time) string {
	weekday := int(now.Weekday())
	if weekday == 0 {
		if now.Hour() >= 18 {
			return now.AddDate(0, 0, 1).Format("2006-01-02")
		}
		weekday = 7
	}

	thisMonday := now.AddDate(0, 0, -(weekday - 1))
	return thisMonday.Format("2006-01-02")
}

func getManagedSignupWeekKey() string {
	return getSignupWeekKeyAt(nowInBotLocation())
}

func isPublicSignupOpen() bool {
	if signupState.ForcedPublicSignupOpen == nil {
		return true
	}

	return *signupState.ForcedPublicSignupOpen
}

func canUserManageSignup(userID string) bool {
	return isAdminUser(userID) || isPublicSignupOpen()
}

func canUserOpenSignupPanel(userID string) bool {
	return isAdminUser(userID)
}

func setForcedPublicSignupOpen(open bool) {
	signupState.ForcedPublicSignupOpen = &open
	saveSignupScheduleState()
}

func buildSignupComponents() []discordgo.MessageComponent {
	return []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{Label: dayLabels["day_mon"], Style: discordgo.PrimaryButton, CustomID: "day_mon"},
				discordgo.Button{Label: dayLabels["day_tue"], Style: discordgo.PrimaryButton, CustomID: "day_tue"},
				discordgo.Button{Label: dayLabels["day_wed"], Style: discordgo.PrimaryButton, CustomID: "day_wed"},
				discordgo.Button{Label: dayLabels["day_thu"], Style: discordgo.PrimaryButton, CustomID: "day_thu"},
				discordgo.Button{Label: dayLabels["day_fri"], Style: discordgo.PrimaryButton, CustomID: "day_fri"},
			},
		},
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{Label: dayLabels["day_sat"], Style: discordgo.SuccessButton, CustomID: "day_sat"},
				discordgo.Button{Label: dayLabels["day_sun"], Style: discordgo.SuccessButton, CustomID: "day_sun"},
			},
		},
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{Label: "我要打全部", Style: discordgo.DangerButton, CustomID: "day_all"},
			},
		},
	}
}

func buildSignupPanelContent(weekKey string) string {
	return fmt.Sprintf(
		"%s 報名\n\n每一天最多 %d 人，額滿後停止該日報名。\n請選擇你要報名的日期：\n\n%s",
		getWeekRangeText(weekKey),
		maxSignupUsersPerDay,
		buildSignupSummary(weekKey),
	)
}

func buildSignupClosedContent() string {
	mentions := collectAllAdminMentions()
	if len(mentions) == 0 {
		return "本週報名已關閉，需要新增請聯絡管理員。"
	}

	return fmt.Sprintf("本週報名已關閉，需要新增請聯絡管理員 %s", joinOrNone(mentions))
}

func sendSignupPanelToChannelOrConfiguredGuild(dg *discordgo.Session, guildID string, fallbackChannelID string, weekKey string) error {
	if fallbackChannelID == "" {
		return fmt.Errorf("missing channel for guild %s", guildID)
	}

	_, err := dg.ChannelMessageSendComplex(fallbackChannelID, &discordgo.MessageSend{
		Content:    buildSignupPanelContent(weekKey),
		Components: buildSignupComponents(),
	})
	return err
}

func sendSignupClosedNoticeToChannelOrConfiguredGuild(dg *discordgo.Session, guildID string, fallbackChannelID string) error {
	if fallbackChannelID == "" {
		return fmt.Errorf("missing channel for guild %s", guildID)
	}

	_, err := dg.ChannelMessageSend(fallbackChannelID, buildSignupClosedContent())
	return err
}

func collectAllAdminMentions() []string {
	seen := map[string]bool{}
	result := []string{}

	for _, userID := range appConfig.AdminUserIDs {
		if seen[userID] {
			continue
		}
		seen[userID] = true
		result = append(result, mentionUser(userID))
	}

	for userID, enabled := range adminState.AdminUsers {
		if !enabled || seen[userID] {
			continue
		}
		seen[userID] = true
		result = append(result, mentionUser(userID))
	}

	return result
}
