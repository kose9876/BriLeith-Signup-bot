package main

import "github.com/bwmarrin/discordgo"

func registerAdminCommands(dg *discordgo.Session, cfg Config) error {
	return registerCommands(dg, cfg, buildAdminCommands())
}

func buildAdminCommands() []*discordgo.ApplicationCommand {
	return []*discordgo.ApplicationCommand{
		{
			Name:        "admin_profile",
			Description: "查看指定玩家的 profile 與報名資訊",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "player",
					Description: "玩家 ID、@mention、遊戲名或顯示名",
					Required:    true,
				},
			},
		},
		{
			Name:        "admin_list",
			Description: "查看管理總覽",
		},
		{
			Name:        "admin_list_players",
			Description: "列出所有已註冊玩家",
		},
		{
			Name:        "admin_addplayer",
			Description: "將新玩家加入名單",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "user_id",
					Description: "Discord user ID 或 @mention",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "game_name",
					Description: "遊戲名稱",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "display_name",
					Description: "顯示名稱，離開 server 的玩家建議填寫",
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "group",
					Description: "群組",
					Choices:     buildGroupChoices(),
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "main_role",
					Description: "主職",
					Choices:     buildRoleChoices(false),
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "sub_role",
					Description: "副職",
					Choices:     buildRoleChoices(true),
				},
				{
					Type:        discordgo.ApplicationCommandOptionBoolean,
					Name:        "has_cape",
					Description: "是否有破袍",
				},
			},
		},
		{
			Name:        "admin_setrole",
			Description: "編輯現有成員資料",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "player",
					Description: "玩家 ID、@mention、遊戲名或顯示名",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "game_name",
					Description: "新的遊戲名稱",
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "main_role",
					Description: "新的主職",
					Choices:     buildRoleChoices(false),
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "sub_role",
					Description: "新的副職",
					Choices:     buildRoleChoices(true),
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "group",
					Description: "新的群組",
					Choices:     buildGroupChoices(),
				},
				{
					Type:        discordgo.ApplicationCommandOptionBoolean,
					Name:        "has_cape",
					Description: "是否有破袍",
				},
			},
		},
		{
			Name:        "admin_setgamename",
			Description: "只更新既有成員的遊戲名稱",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "player",
					Description: "玩家 ID、@mention、遊戲名或顯示名",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "game_name",
					Description: "新的遊戲名稱",
					Required:    true,
				},
			},
		},
		{
			Name:        "admin_removeplayer",
			Description: "移除已註冊玩家",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "player",
					Description: "玩家 ID、@mention、遊戲名或顯示名",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionBoolean,
					Name:        "remove_signup",
					Description: "是否一併移除本週報名",
					Required:    true,
				},
			},
		},
		{
			Name:        "admin_grant",
			Description: "將已註冊玩家加入管理員",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "player",
					Description: "玩家 ID、@mention、遊戲名或顯示名",
					Required:    true,
				},
			},
		},
		{
			Name:        "admin_revoke",
			Description: "移除動態管理員權限",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "player",
					Description: "玩家 ID、@mention、遊戲名或顯示名",
					Required:    true,
				},
			},
		},
		{
			Name:        "admin_signup",
			Description: "手動幫指定玩家報名某一天",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "player",
					Description: "玩家 ID、@mention、遊戲名或顯示名",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "day",
					Description: "要報名的日期",
					Required:    true,
					Choices:     buildDayChoices(),
				},
			},
		},
		{
			Name:        "admin_unsignup",
			Description: "手動取消指定玩家某一天的報名",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "player",
					Description: "玩家 ID、@mention、遊戲名或顯示名",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "day",
					Description: "要取消的日期",
					Required:    true,
					Choices:     buildDayChoices(),
				},
			},
		},
		{
			Name:        "admin_test_signup",
			Description: "手動幫指定玩家加入測試報名某一天",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "player",
					Description: "玩家 ID、@mention、遊戲名或顯示名",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "day",
					Description: "要報名的日期",
					Required:    true,
					Choices:     buildDayChoices(),
				},
			},
		},
		{
			Name:        "admin_test_unsignup",
			Description: "手動取消指定玩家某一天的測試報名",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "player",
					Description: "玩家 ID、@mention、遊戲名或顯示名",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "day",
					Description: "要取消的日期",
					Required:    true,
					Choices:     buildDayChoices(),
				},
			},
		},
		{
			Name:        "admin_test_signup_post",
			Description: "測試手動發送報名表到目前 guild 的設定頻道",
		},
		{
			Name:        "admin_test_summary",
			Description: "查看測試報名的輸出結果",
		},
		{
			Name:        "admin_summary_image",
			Description: "輸出正式報名的表格圖片",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "day",
					Description: "要輸出的日期",
					Required:    true,
					Choices:     buildDayChoices(),
				},
			},
		},
		{
			Name:        "admin_test_summary_image",
			Description: "輸出測試報名的表格圖片",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "day",
					Description: "要輸出的日期",
					Required:    true,
					Choices:     buildDayChoices(),
				},
			},
		},
		{
			Name:        "admin_signup_access",
			Description: "設定玩家是否能自行報名",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "player",
					Description: "玩家 ID、@mention、遊戲名或顯示名",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionBoolean,
					Name:        "blocked",
					Description: "true 為禁止，false 為解除",
					Required:    true,
				},
			},
		},
	}
}

func buildRoleChoices(includeNone bool) []*discordgo.ApplicationCommandOptionChoice {
	choices := []*discordgo.ApplicationCommandOptionChoice{}
	if includeNone {
		choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
			Name:  "none",
			Value: "none",
		})
	}

	choices = append(choices,
		&discordgo.ApplicationCommandOptionChoice{Name: "tank", Value: "tank"},
		&discordgo.ApplicationCommandOptionChoice{Name: "healer", Value: "healer"},
		&discordgo.ApplicationCommandOptionChoice{Name: "dps", Value: "dps"},
	)

	return choices
}

func buildGroupChoices() []*discordgo.ApplicationCommandOptionChoice {
	return []*discordgo.ApplicationCommandOptionChoice{
		{Name: "experienced", Value: "experienced"},
		{Name: "newbie", Value: "newbie"},
	}
}

func buildDayChoices() []*discordgo.ApplicationCommandOptionChoice {
	return []*discordgo.ApplicationCommandOptionChoice{
		{Name: "Monday", Value: "day_mon"},
		{Name: "Tuesday", Value: "day_tue"},
		{Name: "Wednesday", Value: "day_wed"},
		{Name: "Thursday", Value: "day_thu"},
		{Name: "Friday", Value: "day_fri"},
		{Name: "Saturday", Value: "day_sat"},
		{Name: "Sunday", Value: "day_sun"},
	}
}

func findOption(options []*discordgo.ApplicationCommandInteractionDataOption, name string) *discordgo.ApplicationCommandInteractionDataOption {
	for _, option := range options {
		if option.Name == name {
			return option
		}
	}
	return nil
}
