package main

import "github.com/bwmarrin/discordgo"

func registerBaseCommands(dg *discordgo.Session, cfg Config) error {
	return registerCommands(dg, cfg, buildBaseCommands())
}

func buildBaseCommands() []*discordgo.ApplicationCommand {
	return []*discordgo.ApplicationCommand{
		{
			Name:        "setgamename",
			Description: "設定自己的遊戲名稱",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "name",
					Description: "遊戲名稱",
					Required:    true,
				},
			},
		},
		{
			Name:        "signup",
			Description: "開啟報名面板",
		},
		{
			Name:        "whatrole",
			Description: "查看指定玩家的職業設定",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionUser,
					Name:        "user",
					Description: "要查看的玩家",
					Required:    true,
				},
			},
		},
		{
			Name:        "setrole",
			Description: "設定自己的主副職與破袍",
		},
		{
			Name:        "summary",
			Description: "查看本週分配摘要",
		},
		{
			Name:        "help",
			Description: "查看指令說明",
		},
	}
}
