package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/bwmarrin/discordgo"
)

type UserProfile struct {
	DisplayName string `json:"display_name"`
	GameName    string `json:"game_name"`
	MainRole    string `json:"main_role"`
	SubRole     string `json:"sub_role,omitempty"`
	Group       string `json:"group"`
	HasCape     bool   `json:"has_cape"`
}

var userProfiles = map[string]UserProfile{}

var roleLabels = map[string]string{
	"tank":   "坦克",
	"healer": "補師",
	"dps":    "輸出",
}

var groupLabels = map[string]string{
	"experienced": "熟練組",
	"newbie":      "新手組",
}

func updateUserCape(userID string, username string, hasCape bool) {
	profile := userProfiles[userID]
	profile.DisplayName = username
	profile.HasCape = hasCape
	userProfiles[userID] = profile
	saveProfiles()
}

func updateGameName(userID string, username string, gameName string) {
	profile := userProfiles[userID]
	profile.DisplayName = username
	profile.GameName = gameName
	userProfiles[userID] = profile
	saveProfiles()
}

func updateUserGroup(userID string, username string, group string) {
	profile := userProfiles[userID]
	profile.DisplayName = username
	profile.Group = group
	userProfiles[userID] = profile
	saveProfiles()
}

func getRoleLabel(role string) string {
	if role == "" {
		return "未設定"
	}

	if label, exists := roleLabels[role]; exists {
		return label
	}

	return role
}

func getGroupLabel(group string) string {
	if group == "" {
		return "未設定"
	}

	if label, exists := groupLabels[group]; exists {
		return label
	}

	return group
}

func loadProfiles() {
	if err := readJSONFile("profiles.json", &userProfiles); err != nil {
		fmt.Println("load profiles.json failed:", err)
		userProfiles = map[string]UserProfile{}
		return
	}
}

func saveProfiles() {
	data, err := json.MarshalIndent(userProfiles, "", "  ")
	if err != nil {
		fmt.Println("marshal profiles.json failed:", err)
		return
	}

	if err := os.WriteFile("profiles.json", data, 0644); err != nil {
		fmt.Println("write profiles.json failed:", err)
	}
}

func buildSetRoleComponents() []discordgo.MessageComponent {
	return []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{Label: "主職：坦克", Style: discordgo.PrimaryButton, CustomID: "role_main_tank"},
				discordgo.Button{Label: "主職：補師", Style: discordgo.PrimaryButton, CustomID: "role_main_healer"},
				discordgo.Button{Label: "主職：輸出", Style: discordgo.PrimaryButton, CustomID: "role_main_dps"},
			},
		},
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{Label: "副職：無", Style: discordgo.SuccessButton, CustomID: "role_sub_none"},
				discordgo.Button{Label: "副職：坦克", Style: discordgo.SuccessButton, CustomID: "role_sub_tank"},
				discordgo.Button{Label: "副職：補師", Style: discordgo.SuccessButton, CustomID: "role_sub_healer"},
				discordgo.Button{Label: "副職：輸出", Style: discordgo.SuccessButton, CustomID: "role_sub_dps"},
			},
		},
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{Label: "群組：熟練組", Style: discordgo.SecondaryButton, CustomID: "group_experienced"},
				discordgo.Button{Label: "群組：新手組", Style: discordgo.SecondaryButton, CustomID: "group_newbie"},
			},
		},
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{Label: "有破袍", Style: discordgo.SecondaryButton, CustomID: "cape_yes"},
				discordgo.Button{Label: "沒破袍", Style: discordgo.SecondaryButton, CustomID: "cape_no"},
			},
		},
	}
}

func buildSetRoleContent(userID string, username string) string {
	profile := userProfiles[userID]

	capeText := "未設定"
	switch {
	case profile.HasCape:
		capeText = "有破袍"
	case !profile.HasCape:
		capeText = "沒破袍"
	}

	return fmt.Sprintf(
		"%s 的職業設定\n\n主職：%s\n副職：%s\n群組：%s\n破袍：%s\n\n請使用下方按鈕更新設定",
		username,
		getRoleLabel(profile.MainRole),
		getRoleLabel(profile.SubRole),
		getGroupLabel(profile.Group),
		capeText,
	)
}

func updateUserRole(userID string, username string, roleType string, roleValue string) error {
	profile := userProfiles[userID]
	profile.DisplayName = username

	switch roleType {
	case "main":
		profile.MainRole = roleValue
		if profile.SubRole == roleValue {
			profile.SubRole = ""
		}
	case "sub":
		if roleValue == "none" {
			profile.SubRole = ""
		} else {
			if roleValue == profile.MainRole {
				return fmt.Errorf("副職不能和主職相同")
			}
			profile.SubRole = roleValue
		}
	}

	userProfiles[userID] = profile
	saveProfiles()
	return nil
}
