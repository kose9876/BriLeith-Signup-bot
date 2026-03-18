package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type AdminState struct {
	AdminUsers         map[string]bool `json:"admin_users"`
	BlockedSignupUsers map[string]bool `json:"blocked_signup_users"`
}

var adminState = AdminState{
	AdminUsers:         map[string]bool{},
	BlockedSignupUsers: map[string]bool{},
}

func loadAdminState() {
	data, err := os.ReadFile("admin_state.json")
	if err != nil {
		adminState = AdminState{
			AdminUsers:         map[string]bool{},
			BlockedSignupUsers: map[string]bool{},
		}
		return
	}

	err = json.Unmarshal(data, &adminState)
	if err != nil {
		fmt.Println("讀取 admin_state.json 失敗:", err)
		adminState = AdminState{
			AdminUsers:         map[string]bool{},
			BlockedSignupUsers: map[string]bool{},
		}
		return
	}

	if adminState.AdminUsers == nil {
		adminState.AdminUsers = map[string]bool{}
	}
	if adminState.BlockedSignupUsers == nil {
		adminState.BlockedSignupUsers = map[string]bool{}
	}
}

func saveAdminState() {
	data, err := json.MarshalIndent(adminState, "", "  ")
	if err != nil {
		fmt.Println("轉換 admin_state.json 失敗:", err)
		return
	}

	err = os.WriteFile("admin_state.json", data, 0644)
	if err != nil {
		fmt.Println("寫入 admin_state.json 失敗:", err)
	}
}

func isAdminUser(userID string) bool {
	for _, adminUserID := range appConfig.AdminUserIDs {
		if adminUserID == userID {
			return true
		}
	}

	return adminState.AdminUsers[userID]
}

func isSignupBlocked(userID string) bool {
	return adminState.BlockedSignupUsers[userID]
}

func registerAdminCommands(dg *discordgo.Session, cfg Config) error {
	commands := []*discordgo.ApplicationCommand{
		{
			Name:        "admin_profile",
			Description: "查看指定玩家的 profile 與報名資訊",
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
					Type:        discordgo.ApplicationCommandOptionUser,
					Name:        "user",
					Description: "要加入的玩家",
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
					Type:        discordgo.ApplicationCommandOptionUser,
					Name:        "user",
					Description: "要編輯的玩家",
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
					Type:        discordgo.ApplicationCommandOptionUser,
					Name:        "user",
					Description: "要更新的玩家",
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
					Type:        discordgo.ApplicationCommandOptionUser,
					Name:        "user",
					Description: "要移除的玩家",
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
					Type:        discordgo.ApplicationCommandOptionUser,
					Name:        "user",
					Description: "要授權的玩家",
					Required:    true,
				},
			},
		},
		{
			Name:        "admin_revoke",
			Description: "移除動態管理員權限",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionUser,
					Name:        "user",
					Description: "要移除權限的玩家",
					Required:    true,
				},
			},
		},
		{
			Name:        "admin_signup",
			Description: "手動幫指定玩家報名某一天",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionUser,
					Name:        "user",
					Description: "要代報名的玩家",
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
					Type:        discordgo.ApplicationCommandOptionUser,
					Name:        "user",
					Description: "要取消報名的玩家",
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
			Name:        "admin_test_signup_window",
			Description: "測試手動開關一般玩家報名",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionBoolean,
					Name:        "open",
					Description: "true 開啟，false 關閉",
					Required:    true,
				},
			},
		},
		{
			Name:        "admin_test_signup_close_notice",
			Description: "測試手動發送報名關閉通知到目前 guild 的設定頻道",
		},
		{
			Name:        "admin_signup_access",
			Description: "設定玩家是否能自行報名",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionUser,
					Name:        "user",
					Description: "要調整的玩家",
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

	for _, guildID := range cfg.GuildIDs {
		for _, command := range commands {
			_, err := dg.ApplicationCommandCreate(cfg.ApplicationID, guildID, command)
			if err != nil {
				return fmt.Errorf("guild %s command %s: %w", guildID, command.Name, err)
			}
		}
	}

	return nil
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

func respondEphemeral(s *discordgo.Session, i *discordgo.InteractionCreate, content string) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		fmt.Println("回覆互動訊息失敗:", err)
	}
}

func requireAdmin(s *discordgo.Session, i *discordgo.InteractionCreate) bool {
	if i.Member == nil || i.Member.User == nil || !isAdminUser(i.Member.User.ID) {
		respondEphemeral(s, i, "你不是管理員，無法使用這個指令。")
		return false
	}

	return true
}

func handleAdminCommand(s *discordgo.Session, i *discordgo.InteractionCreate) bool {
	if i.Type != discordgo.InteractionApplicationCommand {
		return false
	}

	commandName := i.ApplicationCommandData().Name
	if !strings.HasPrefix(commandName, "admin_") {
		return false
	}

	if !requireAdmin(s, i) {
		return true
	}

	switch commandName {
	case "admin_profile":
		handleAdminProfileCommand(s, i)
	case "admin_list":
		handleAdminListCommand(s, i)
	case "admin_list_players":
		handleAdminListPlayersCommand(s, i)
	case "admin_addplayer":
		handleAdminAddPlayerCommand(s, i)
	case "admin_setrole":
		handleAdminSetRoleCommand(s, i)
	case "admin_setgamename":
		handleAdminSetGameNameCommand(s, i)
	case "admin_removeplayer":
		handleAdminRemovePlayerCommand(s, i)
	case "admin_grant":
		handleAdminGrantCommand(s, i)
	case "admin_revoke":
		handleAdminRevokeCommand(s, i)
	case "admin_signup":
		handleAdminSignupCommand(s, i)
	case "admin_unsignup":
		handleAdminUnsignupCommand(s, i)
	case "admin_test_signup_post":
		handleAdminTestSignupPanelCommand(s, i)
	case "admin_test_summary":
		handleAdminTestSummaryCommand(s, i)
	case "admin_test_signup_window":
		handleAdminTestSignupWindowCommand(s, i)
	case "admin_test_signup_close_notice":
		handleAdminTestSignupCloseNoticePreviewCommand(s, i)
	case "admin_signup_access":
		handleAdminSignupAccessCommand(s, i)
	default:
		respondEphemeral(s, i, "未知的管理指令。")
	}

	return true
}

func handleAdminProfileCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options
	userOption := findOption(options, "user")
	if userOption == nil {
		respondEphemeral(s, i, "缺少 user 參數。")
		return
	}

	user := userOption.UserValue(s)
	profile, exists := userProfiles[user.ID]
	weekKey := getManagedSignupWeekKey()

	currentWeekDays := []string{}
	if weeklySignups[weekKey] != nil {
		currentWeekDays = weeklySignups[weekKey][user.ID]
	}

	if !exists {
		respondEphemeral(s, i, fmt.Sprintf(
			"玩家: <@%s>\n目前沒有 profile。\n本週報名: %s\n自行報名權限: %s",
			user.ID,
			formatSignupDays(currentWeekDays),
			getSignupAccessText(user.ID),
		))
		return
	}

	content := fmt.Sprintf(
		"玩家: <@%s>\n顯示名稱: %s\n遊戲名稱: %s\n主職: %s\n副職: %s\n群組: %s\n破袍: %s\n本週報名: %s\n自行報名權限: %s",
		user.ID,
		emptyFallback(profile.DisplayName),
		emptyFallback(profile.GameName),
		getRoleLabel(profile.MainRole),
		getRoleLabel(profile.SubRole),
		emptyFallback(profile.Group),
		boolText(profile.HasCape, "有", "沒有"),
		formatSignupDays(currentWeekDays),
		getSignupAccessText(user.ID),
	)

	respondEphemeral(s, i, content)
}

func handleAdminListCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	rootAdmins := make([]string, 0, len(appConfig.AdminUserIDs))
	for _, userID := range appConfig.AdminUserIDs {
		rootAdmins = append(rootAdmins, mentionUser(userID))
	}

	dynamicAdmins := collectEnabledUserMentions(adminState.AdminUsers)
	blockedUsers := collectEnabledUserMentions(adminState.BlockedSignupUsers)

	content := fmt.Sprintf(
		"管理總覽\n固定管理員: %s\n動態管理員: %s\n禁止自行報名: %s\n已註冊玩家數: %d",
		joinOrNone(rootAdmins),
		joinOrNone(dynamicAdmins),
		joinOrNone(blockedUsers),
		len(userProfiles),
	)

	respondEphemeral(s, i, content)
}

func handleAdminListPlayersCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if len(userProfiles) == 0 {
		respondEphemeral(s, i, "目前沒有已註冊玩家。")
		return
	}

	type playerLine struct {
		SortKey string
		Text    string
	}

	lines := make([]playerLine, 0, len(userProfiles))
	for userID, profile := range userProfiles {
		sortKey := strings.ToLower(emptyFallback(profile.GameName) + emptyFallback(profile.DisplayName) + userID)
		text := fmt.Sprintf(
			"<@%s> | %s | 主職:%s | 副職:%s | 群組:%s",
			userID,
			emptyFallback(profile.GameName),
			getRoleLabel(profile.MainRole),
			getRoleLabel(profile.SubRole),
			emptyFallback(profile.Group),
		)
		lines = append(lines, playerLine{SortKey: sortKey, Text: text})
	}

	sort.Slice(lines, func(i, j int) bool {
		return lines[i].SortKey < lines[j].SortKey
	})

	texts := make([]string, 0, len(lines))
	for _, line := range lines {
		texts = append(texts, line.Text)
	}

	respondEphemeral(s, i, "已註冊玩家清單\n"+strings.Join(texts, "\n"))
}

func handleAdminAddPlayerCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options

	userOption := findOption(options, "user")
	gameNameOption := findOption(options, "game_name")
	if userOption == nil || gameNameOption == nil {
		respondEphemeral(s, i, "缺少必要參數。")
		return
	}

	user := userOption.UserValue(s)

	profile := userProfiles[user.ID]
	profile.DisplayName = user.Username
	profile.GameName = gameNameOption.StringValue()

	if option := findOption(options, "group"); option != nil {
		profile.Group = option.StringValue()
	}
	if option := findOption(options, "main_role"); option != nil {
		profile.MainRole = option.StringValue()
		if profile.SubRole == profile.MainRole {
			profile.SubRole = ""
		}
	}
	if option := findOption(options, "sub_role"); option != nil {
		subRole := option.StringValue()
		if subRole == "none" {
			subRole = ""
		}
		if subRole != "" && subRole == profile.MainRole {
			respondEphemeral(s, i, "副職不能和主職相同。")
			return
		}
		profile.SubRole = subRole
	}
	if option := findOption(options, "has_cape"); option != nil {
		profile.HasCape = option.BoolValue()
	}

	userProfiles[user.ID] = profile
	saveProfiles()

	respondEphemeral(s, i, fmt.Sprintf(
		"已將 <@%s> 加入或更新到名單。\n遊戲名稱: %s\n主職: %s\n副職: %s",
		user.ID,
		emptyFallback(profile.GameName),
		getRoleLabel(profile.MainRole),
		getRoleLabel(profile.SubRole),
	))
}

func handleAdminSetRoleCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options

	userOption := findOption(options, "user")
	if userOption == nil {
		respondEphemeral(s, i, "缺少必要參數。")
		return
	}

	user := userOption.UserValue(s)
	profile, exists := userProfiles[user.ID]
	if !exists {
		respondEphemeral(s, i, "這個玩家還沒有資料，請先使用 /admin_addplayer 建立名單。")
		return
	}

	profile.DisplayName = user.Username

	if option := findOption(options, "game_name"); option != nil {
		profile.GameName = option.StringValue()
	}
	if option := findOption(options, "main_role"); option != nil {
		profile.MainRole = option.StringValue()
		if profile.SubRole == profile.MainRole {
			profile.SubRole = ""
		}
	}
	if option := findOption(options, "sub_role"); option != nil {
		subRole := option.StringValue()
		if subRole == "none" {
			subRole = ""
		}
		if subRole != "" && subRole == profile.MainRole {
			respondEphemeral(s, i, "副職不能和主職相同。")
			return
		}
		profile.SubRole = subRole
	}
	if option := findOption(options, "group"); option != nil {
		profile.Group = option.StringValue()
	}
	if option := findOption(options, "has_cape"); option != nil {
		profile.HasCape = option.BoolValue()
	}

	userProfiles[user.ID] = profile
	saveProfiles()

	respondEphemeral(s, i, fmt.Sprintf(
		"已更新 <@%s> 的資料。\n遊戲名稱: %s\n主職: %s\n副職: %s\n群組: %s\n破袍: %s",
		user.ID,
		emptyFallback(profile.GameName),
		getRoleLabel(profile.MainRole),
		getRoleLabel(profile.SubRole),
		emptyFallback(profile.Group),
		boolText(profile.HasCape, "有", "沒有"),
	))
}

func handleAdminSetGameNameCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options

	userOption := findOption(options, "user")
	gameNameOption := findOption(options, "game_name")
	if userOption == nil || gameNameOption == nil {
		respondEphemeral(s, i, "缺少必要參數。")
		return
	}

	user := userOption.UserValue(s)
	profile, exists := userProfiles[user.ID]
	if !exists {
		respondEphemeral(s, i, "這個玩家還沒有資料，請先使用 /admin_addplayer 建立名單。")
		return
	}

	profile.DisplayName = user.Username
	profile.GameName = gameNameOption.StringValue()
	userProfiles[user.ID] = profile
	saveProfiles()

	respondEphemeral(s, i, fmt.Sprintf(
		"已更新 <@%s> 的遊戲名稱為 %s。",
		user.ID,
		emptyFallback(profile.GameName),
	))
}

func handleAdminRemovePlayerCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options

	userOption := findOption(options, "user")
	removeSignupOption := findOption(options, "remove_signup")
	if userOption == nil || removeSignupOption == nil {
		respondEphemeral(s, i, "缺少必要參數。")
		return
	}

	user := userOption.UserValue(s)
	if _, exists := userProfiles[user.ID]; !exists {
		respondEphemeral(s, i, "這個玩家目前不在名單內。")
		return
	}

	delete(userProfiles, user.ID)
	saveProfiles()

	delete(adminState.AdminUsers, user.ID)
	delete(adminState.BlockedSignupUsers, user.ID)
	saveAdminState()

	if removeSignupOption.BoolValue() {
		removeUserSignupWeek(getManagedSignupWeekKey(), user.ID)
	}

	respondEphemeral(s, i, fmt.Sprintf(
		"已從名單移除 <@%s>。%s",
		user.ID,
		boolText(removeSignupOption.BoolValue(), "本週報名也已刪除。", "本週報名保留。"),
	))
}

func handleAdminGrantCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options

	userOption := findOption(options, "user")
	if userOption == nil {
		respondEphemeral(s, i, "缺少必要參數。")
		return
	}

	user := userOption.UserValue(s)
	if _, exists := userProfiles[user.ID]; !exists {
		respondEphemeral(s, i, "這個玩家還沒有註冊資料，請先建立名單。")
		return
	}

	adminState.AdminUsers[user.ID] = true
	saveAdminState()

	respondEphemeral(s, i, fmt.Sprintf(
		"已將 <@%s> 加入管理員名單。",
		user.ID,
	))
}

func handleAdminRevokeCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options

	userOption := findOption(options, "user")
	if userOption == nil {
		respondEphemeral(s, i, "缺少必要參數。")
		return
	}

	user := userOption.UserValue(s)

	for _, rootAdminID := range appConfig.AdminUserIDs {
		if rootAdminID == user.ID {
			respondEphemeral(s, i, "這個使用者是 config.json 的固定管理員，不能用這個指令移除。")
			return
		}
	}

	if !adminState.AdminUsers[user.ID] {
		respondEphemeral(s, i, "這個使用者目前不是動態管理員。")
		return
	}

	delete(adminState.AdminUsers, user.ID)
	saveAdminState()

	respondEphemeral(s, i, fmt.Sprintf(
		"已移除 <@%s> 的管理員權限。",
		user.ID,
	))
}

func handleAdminSignupCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options

	userOption := findOption(options, "user")
	dayOption := findOption(options, "day")
	if userOption == nil || dayOption == nil {
		respondEphemeral(s, i, "缺少必要參數。")
		return
	}

	user := userOption.UserValue(s)
	day := dayOption.StringValue()

	addUserSignupDay(getManagedSignupWeekKey(), user.ID, day)

	respondEphemeral(s, i, fmt.Sprintf(
		"已幫 <@%s> 手動報名 %s。",
		user.ID,
		getDayLabel(day),
	))
}

func handleAdminUnsignupCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options

	userOption := findOption(options, "user")
	dayOption := findOption(options, "day")
	if userOption == nil || dayOption == nil {
		respondEphemeral(s, i, "缺少必要參數。")
		return
	}

	user := userOption.UserValue(s)
	day := dayOption.StringValue()

	if !removeUserSignupDay(getManagedSignupWeekKey(), user.ID, day) {
		respondEphemeral(s, i, fmt.Sprintf(
			"<@%s> 本週沒有報名 %s。",
			user.ID,
			getDayLabel(day),
		))
		return
	}

	respondEphemeral(s, i, fmt.Sprintf(
		"已取消 <@%s> 的 %s 報名。",
		user.ID,
		getDayLabel(day),
	))
}

func handleAdminTestSignupPostCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	guildID := i.GuildID
	if guildID == "" {
		respondEphemeral(s, i, "這個指令只能在伺服器內使用。")
		return
	}

	weekKey := getSignupWeekKeyAt(nowInBotLocation())
	if err := sendSignupPanelToChannelOrConfiguredGuild(s, guildID, i.ChannelID, weekKey); err != nil {
		respondEphemeral(s, i, "手動發送報名表失敗: "+err.Error())
		return
	}

	respondEphemeral(s, i, fmt.Sprintf(
		"已手動發送 %s 的報名表到目前 guild 設定頻道。",
		getWeekRangeText(weekKey),
	))
}

func handleAdminTestSignupWindowCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options
	openOption := findOption(options, "open")
	if openOption == nil {
		respondEphemeral(s, i, "缺少必要參數。")
		return
	}

	open := openOption.BoolValue()
	setForcedPublicSignupOpen(open)

	if open {
		respondEphemeral(s, i, "已強制開啟一般玩家報名。")
		return
	}

	respondEphemeral(s, i, "已強制關閉一般玩家報名。")
}

func handleAdminTestSignupCloseNoticeCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	guildID := i.GuildID
	if guildID == "" {
		respondEphemeral(s, i, "這個指令只能在伺服器內使用。")
		return
	}

	if err := sendSignupClosedNoticeToChannelOrConfiguredGuild(s, guildID, i.ChannelID); err != nil {
		respondEphemeral(s, i, "手動發送關閉通知失敗: "+err.Error())
		return
	}

	respondEphemeral(s, i, "已手動發送報名關閉通知到目前 guild 設定頻道。")
}

func handleAdminSignupAccessCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options

	userOption := findOption(options, "user")
	blockedOption := findOption(options, "blocked")
	if userOption == nil || blockedOption == nil {
		respondEphemeral(s, i, "缺少必要參數。")
		return
	}

	user := userOption.UserValue(s)
	blocked := blockedOption.BoolValue()

	if blocked {
		adminState.BlockedSignupUsers[user.ID] = true
	} else {
		delete(adminState.BlockedSignupUsers, user.ID)
	}
	saveAdminState()

	respondEphemeral(s, i, fmt.Sprintf(
		"已更新 <@%s> 的報名權限: %s",
		user.ID,
		getSignupAccessText(user.ID),
	))
}

func addUserSignupDay(weekKey string, userID string, day string) {
	if weeklySignups[weekKey] == nil {
		weeklySignups[weekKey] = map[string][]string{}
	}

	for _, existingDay := range weeklySignups[weekKey][userID] {
		if existingDay == day {
			return
		}
	}

	weeklySignups[weekKey][userID] = append(weeklySignups[weekKey][userID], day)
	sortUserDays(weeklySignups[weekKey][userID])
	saveSignups()
}

func removeUserSignupDay(weekKey string, userID string, day string) bool {
	if weeklySignups[weekKey] == nil {
		return false
	}

	days := weeklySignups[weekKey][userID]
	for idx, existingDay := range days {
		if existingDay != day {
			continue
		}

		updated := append([]string{}, days[:idx]...)
		updated = append(updated, days[idx+1:]...)
		if len(updated) == 0 {
			delete(weeklySignups[weekKey], userID)
		} else {
			weeklySignups[weekKey][userID] = updated
		}
		saveSignups()
		return true
	}

	return false
}

func removeUserSignupWeek(weekKey string, userID string) {
	if weeklySignups[weekKey] == nil {
		return
	}

	delete(weeklySignups[weekKey], userID)
	saveSignups()
}

func formatSignupDays(days []string) string {
	if len(days) == 0 {
		return "尚未報名"
	}

	labels := make([]string, 0, len(days))
	for _, day := range days {
		labels = append(labels, getDayLabel(day))
	}

	return strings.Join(labels, "、")
}

func getSignupAccessText(userID string) string {
	if isSignupBlocked(userID) {
		return "已禁止自行報名"
	}

	return "可自行報名"
}

func emptyFallback(value string) string {
	if strings.TrimSpace(value) == "" {
		return "未設定"
	}

	return value
}

func boolText(value bool, yes string, no string) string {
	if value {
		return yes
	}

	return no
}

func getDayLabel(day string) string {
	if label, exists := dayLabels[day]; exists {
		return label
	}

	return day
}

func mentionUser(userID string) string {
	return fmt.Sprintf("<@%s>", userID)
}

func collectEnabledUserMentions(items map[string]bool) []string {
	userIDs := make([]string, 0, len(items))
	for userID, enabled := range items {
		if enabled {
			userIDs = append(userIDs, userID)
		}
	}

	sort.Strings(userIDs)

	result := make([]string, 0, len(userIDs))
	for _, userID := range userIDs {
		result = append(result, mentionUser(userID))
	}

	return result
}

func joinOrNone(items []string) string {
	if len(items) == 0 {
		return "無"
	}

	return strings.Join(items, "、")
}
