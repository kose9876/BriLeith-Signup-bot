package main

import (
	"strings"

	"github.com/bwmarrin/discordgo"
)

func buildHelpText() string {
	playerCommands := []string{
		"/setrole name:<遊戲名稱> 設定自己的遊戲名稱、主職、副職與破袍",
		"/signup 開啟每週報名面板",
		"/summary 查看本週分配摘要",
		"/whatrole user:@玩家 查看指定玩家的職業資料",
	}

	adminCommands := []string{
		"/a_list 查看管理總覽",
		"/a_list_players 列出所有已註冊玩家",
		"/a_profile user:@玩家 查看玩家 profile 與報名資訊",
		"/a_addplayer 建立新玩家資料",
		"/a_setrole 編輯既有玩家的遊戲名稱、主副職、群組、破袍",
		"/a_setgamename 只修改既有玩家的遊戲名稱",
		"/a_removeplayer 移除玩家，可選是否刪除本週報名",
		"/a_signup 手動幫玩家報名某一天",
		"/a_unsignup 手動取消玩家某一天報名",
		"/a_signup_access 設定玩家能否自行報名",
		"/a_grant 將已註冊玩家加入管理員",
		"/a_grant_tester 將已註冊玩家加入測試員",
		"/a_revoke 移除動態管理員權限",
		"/a_revoke_tester 移除測試員權限",
		"/a_summary_image 輸出正式報名的表格圖片",
	}

	testerCommands := []string{
		"/a_profile 查看玩家 profile 與報名資訊",
		"/a_list_players 列出所有已註冊玩家",
		"/t_signup 手動幫玩家加入測試報名某一天",
		"/t_unsignup 手動取消玩家某一天的測試報名",
		"/t_boss3_assign 手動調整測試版三王工作分配",
		"/t_boss3_clear 清除測試版三王工作分配覆寫",
		"/t_signup_post 手動發送測試用報名表",
		"/t_summary 查看測試報名與分配輸出",
		"/t_summary_image 輸出測試報名的表格圖片",
	}

	return "指令說明\n\n一般玩家\n" +
		strings.Join(playerCommands, "\n") +
		"\n\n管理員\n" +
		strings.Join(adminCommands, "\n") +
		"\n\n測試員\n" +
		strings.Join(testerCommands, "\n")
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
