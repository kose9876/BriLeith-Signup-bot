package main

import (
	"strings"

	"github.com/bwmarrin/discordgo"
)

func buildHelpText() string {
	playerCommands := []string{
		"/setgamename name:<遊戲名稱> 設定自己的遊戲名稱",
		"/setrole 設定自己的主職、副職與破袍",
		"/signup 開啟每週報名面板",
		"/summary 查看本週分配摘要",
		"/whatrole user:@玩家 查看指定玩家的職業資料",
	}

	adminCommands := []string{
		"/admin_list 查看管理總覽",
		"/admin_list_players 列出所有已註冊玩家",
		"/admin_profile user:@玩家 查看玩家 profile 與報名資訊",
		"/admin_addplayer 建立新玩家資料",
		"/admin_setrole 編輯既有玩家的遊戲名稱、主副職、群組、破袍",
		"/admin_setgamename 只修改既有玩家的遊戲名稱",
		"/admin_removeplayer 移除玩家，可選是否刪除本週報名",
		"/admin_signup 手動幫玩家報名某一天",
		"/admin_unsignup 手動取消玩家某一天報名",
		"/admin_signup_access 設定玩家能否自行報名",
		"/admin_grant 將已註冊玩家加入管理員",
		"/admin_revoke 移除動態管理員權限",
		"/admin_test_signup_post 手動發送測試用報名表",
		"/admin_test_summary 查看測試報名與分配輸出",
	}

	return "指令說明\n\n一般玩家\n" +
		strings.Join(playerCommands, "\n") +
		"\n\n管理員\n" +
		strings.Join(adminCommands, "\n")
}

func handleHelpCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: buildHelpText(),
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		return
	}
}
