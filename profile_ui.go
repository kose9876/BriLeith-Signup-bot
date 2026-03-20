package main

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

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
