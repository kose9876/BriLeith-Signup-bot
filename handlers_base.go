package main

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

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
			logInteractionCommand(i)
			handleHelpCommand(s, i)
			return
		}

		if handleAdminCommand(s, i) {
			return
		}

		if i.Type == discordgo.InteractionApplicationCommand {
			logInteractionCommand(i)
			handleApplicationCommand(s, i)
			return
		}

		if i.Type == discordgo.InteractionMessageComponent {
			logInteractionComponent(i)
			handleMessageComponent(s, i)
		}
	})
}

func handleApplicationCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.ApplicationCommandData().Name {
	case "signup":
		handleSignupCommand(s, i)
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
	if !canUserOpenSignupPanel(userID) {
		respondEphemeral(s, i, "只有管理員可以開啟報名表。")
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
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		fmt.Println("whatrole command failed:", err)
	}
}

func handleSetRoleCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	userID := i.Member.User.ID
	username := i.Member.User.Username
	gameNameOption := findOption(i.ApplicationCommandData().Options, "name")
	if gameNameOption == nil {
		respondEphemeral(s, i, "缺少 name 參數。")
		return
	}

	updateGameName(userID, username, gameNameOption.StringValue())

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
	content := buildTestFullDaySummaryText(weekKey, dayKey)

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		fmt.Println("test summary detail failed:", err)
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
