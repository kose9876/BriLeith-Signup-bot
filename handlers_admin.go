package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func handleAdminCommand(s *discordgo.Session, i *discordgo.InteractionCreate) bool {
	if i.Type != discordgo.InteractionApplicationCommand {
		return false
	}

	commandName := i.ApplicationCommandData().Name
	if !strings.HasPrefix(commandName, "a_") && !strings.HasPrefix(commandName, "t_") {
		return false
	}

	logInteractionCommand(i)

	if commandName == "a_profile" || commandName == "a_list_players" {
		if !requireAdminOrTester(s, i) {
			return true
		}
	} else if strings.HasPrefix(commandName, "t_") {
		if !requireTester(s, i) {
			return true
		}
	} else if !requireAdmin(s, i) {
		return true
	}

	switch commandName {
	case "a_profile":
		handleAdminProfileCommand(s, i)
	case "a_list":
		handleAdminListCommand(s, i)
	case "a_list_players":
		handleAdminListPlayersCommand(s, i)
	case "a_addplayer":
		handleAdminAddPlayerCommand(s, i)
	case "a_setrole":
		handleAdminSetRoleCommand(s, i)
	case "a_setgamename":
		handleAdminSetGameNameCommand(s, i)
	case "a_removeplayer":
		handleAdminRemovePlayerCommand(s, i)
	case "a_grant":
		handleAdminGrantCommand(s, i)
	case "a_grant_tester":
		handleAdminTesterGrantCommand(s, i)
	case "a_revoke":
		handleAdminRevokeCommand(s, i)
	case "a_revoke_tester":
		handleAdminTesterRevokeCommand(s, i)
	case "a_signup":
		handleAdminSignupCommand(s, i)
	case "a_boss3_assign":
		handleAdminBoss3AssignCommand(s, i)
	case "a_boss3_clear":
		handleAdminBoss3ClearCommand(s, i)
	case "a_unsignup":
		handleAdminUnsignupCommand(s, i)
	case "t_signup":
		handleAdminTestSignupCommand(s, i)
	case "t_unsignup":
		handleAdminTestUnsignupCommand(s, i)
	case "t_signup_post":
		handleAdminTestSignupPanelCommand(s, i)
	case "t_boss3_assign":
		handleAdminTestBoss3AssignCommand(s, i)
	case "t_boss3_clear":
		handleAdminTestBoss3ClearCommand(s, i)
	case "t_summary":
		handleAdminTestSummaryCommand(s, i)
	case "a_summary_image":
		handleAdminSummaryImageCommand(s, i)
	case "t_summary_image":
		handleAdminTestSummaryImageCommand(s, i)
	case "a_signup_access":
		handleAdminSignupAccessCommand(s, i)
	default:
		respondEphemeral(s, i, "未知的管理指令。")
	}

	return true
}

func handleAdminProfileCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options
	playerOption := findOption(options, "player")
	if playerOption == nil {
		respondEphemeral(s, i, "缺少 player 參數。")
		return
	}

	userID, profile, exists, err := resolveRegisteredPlayer(playerOption.StringValue())
	if err != nil {
		respondEphemeral(s, i, err.Error())
		return
	}
	weekKey := getManagedSignupWeekKey()

	currentWeekDays := []string{}
	if weeklySignups[weekKey] != nil {
		currentWeekDays = weeklySignups[weekKey][userID]
	}

	if !exists {
		respondEphemeral(s, i, fmt.Sprintf(
			"玩家: %s\n目前沒有 profile。\n本週報名: %s\n自行報名權限: %s",
			formatPlayerLabel(userID),
			formatSignupDays(currentWeekDays),
			getSignupAccessText(userID),
		))
		return
	}

	content := fmt.Sprintf(
		"玩家: %s\n顯示名稱: %s\n遊戲名稱: %s\n主職: %s\n副職: %s\n群組: %s\n破袍: %s\n本週報名: %s\n自行報名權限: %s",
		formatPlayerLabel(userID),
		emptyFallback(profile.DisplayName),
		emptyFallback(profile.GameName),
		getRoleLabel(profile.MainRole),
		getRoleLabel(profile.SubRole),
		emptyFallback(profile.Group),
		boolText(profile.HasCape, "有", "沒有"),
		formatSignupDays(currentWeekDays),
		getSignupAccessText(userID),
	)

	respondEphemeral(s, i, content)
}

func handleAdminListCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	rootAdmins := make([]string, 0, len(appConfig.AdminUserIDs))
	for _, userID := range appConfig.AdminUserIDs {
		rootAdmins = append(rootAdmins, mentionUser(userID))
	}

	dynamicAdmins := collectEnabledUserMentions(adminState.AdminUsers)
	testers := collectEnabledUserMentions(adminState.TesterUsers)
	blockedUsers := collectEnabledUserMentions(adminState.BlockedSignupUsers)

	content := fmt.Sprintf(
		"管理總覽\n固定管理員: %s\n動態管理員: %s\n測試員: %s\n禁止自行報名: %s\n已註冊玩家數: %d",
		joinOrNone(rootAdmins),
		joinOrNone(dynamicAdmins),
		joinOrNone(testers),
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
			"%s | 主職:%s | 副職:%s | 群組:%s",
			formatPlayerLabel(userID),
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

	userIDOption := findOption(options, "user_id")
	gameNameOption := findOption(options, "game_name")
	if userIDOption == nil || gameNameOption == nil {
		respondEphemeral(s, i, "缺少必要參數。")
		return
	}

	userID, err := normalizeDiscordUserID(userIDOption.StringValue())
	if err != nil {
		respondEphemeral(s, i, err.Error())
		return
	}

	profile := userProfiles[userID]
	if option := findOption(options, "display_name"); option != nil {
		profile.DisplayName = strings.TrimSpace(option.StringValue())
	}
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

	userProfiles[userID] = profile
	saveProfiles()

	respondEphemeral(s, i, fmt.Sprintf(
		"已將 %s 加入或更新到名單。\n遊戲名稱: %s\n主職: %s\n副職: %s",
		formatPlayerLabel(userID),
		emptyFallback(profile.GameName),
		getRoleLabel(profile.MainRole),
		getRoleLabel(profile.SubRole),
	))
}

func handleAdminSetRoleCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options

	playerOption := findOption(options, "player")
	if playerOption == nil {
		respondEphemeral(s, i, "缺少必要參數。")
		return
	}

	userID, profile, exists, err := resolveRegisteredPlayer(playerOption.StringValue())
	if err != nil {
		respondEphemeral(s, i, err.Error())
		return
	}
	if !exists {
		respondEphemeral(s, i, "這個玩家還沒有資料，請先使用 /a_addplayer 建立名單。")
		return
	}

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

	userProfiles[userID] = profile
	saveProfiles()

	respondEphemeral(s, i, fmt.Sprintf(
		"已更新 %s 的資料。\n遊戲名稱: %s\n主職: %s\n副職: %s\n群組: %s\n破袍: %s",
		formatPlayerLabel(userID),
		emptyFallback(profile.GameName),
		getRoleLabel(profile.MainRole),
		getRoleLabel(profile.SubRole),
		emptyFallback(profile.Group),
		boolText(profile.HasCape, "有", "沒有"),
	))
}

func handleAdminSetGameNameCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options

	playerOption := findOption(options, "player")
	gameNameOption := findOption(options, "game_name")
	if playerOption == nil || gameNameOption == nil {
		respondEphemeral(s, i, "缺少必要參數。")
		return
	}

	userID, profile, exists, err := resolveRegisteredPlayer(playerOption.StringValue())
	if err != nil {
		respondEphemeral(s, i, err.Error())
		return
	}
	if !exists {
		respondEphemeral(s, i, "這個玩家還沒有資料，請先使用 /a_addplayer 建立名單。")
		return
	}

	profile.GameName = gameNameOption.StringValue()
	userProfiles[userID] = profile
	saveProfiles()

	respondEphemeral(s, i, fmt.Sprintf(
		"已更新 %s 的遊戲名稱為 %s。",
		formatPlayerLabel(userID),
		emptyFallback(profile.GameName),
	))
}

func handleAdminRemovePlayerCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options

	playerOption := findOption(options, "player")
	removeSignupOption := findOption(options, "remove_signup")
	if playerOption == nil || removeSignupOption == nil {
		respondEphemeral(s, i, "缺少必要參數。")
		return
	}

	userID, _, exists, err := resolveRegisteredPlayer(playerOption.StringValue())
	if err != nil {
		respondEphemeral(s, i, err.Error())
		return
	}
	if !exists {
		respondEphemeral(s, i, "這個玩家目前不在名單內。")
		return
	}

	delete(userProfiles, userID)
	saveProfiles()

	delete(adminState.AdminUsers, userID)
	delete(adminState.TesterUsers, userID)
	delete(adminState.BlockedSignupUsers, userID)
	saveAdminState()

	if removeSignupOption.BoolValue() {
		removeUserSignupWeek(getManagedSignupWeekKey(), userID)
	}

	respondEphemeral(s, i, fmt.Sprintf(
		"已從名單移除 %s。%s",
		formatPlayerLabel(userID),
		boolText(removeSignupOption.BoolValue(), "本週報名也已刪除。", "本週報名保留。"),
	))
}

func handleAdminGrantCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options

	playerOption := findOption(options, "player")
	if playerOption == nil {
		respondEphemeral(s, i, "缺少必要參數。")
		return
	}

	userID, _, exists, err := resolveRegisteredPlayer(playerOption.StringValue())
	if err != nil {
		respondEphemeral(s, i, err.Error())
		return
	}
	if !exists {
		respondEphemeral(s, i, "這個玩家還沒有註冊資料，請先建立名單。")
		return
	}

	adminState.AdminUsers[userID] = true
	saveAdminState()

	respondEphemeral(s, i, fmt.Sprintf(
		"已將 %s 加入管理員名單。",
		formatPlayerLabel(userID),
	))
}

func handleAdminRevokeCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options

	playerOption := findOption(options, "player")
	if playerOption == nil {
		respondEphemeral(s, i, "缺少必要參數。")
		return
	}

	userID, _, exists, err := resolveRegisteredPlayer(playerOption.StringValue())
	if err != nil {
		respondEphemeral(s, i, err.Error())
		return
	}
	if !exists {
		respondEphemeral(s, i, "這個玩家目前不在名單內。")
		return
	}

	for _, rootAdminID := range appConfig.AdminUserIDs {
		if rootAdminID == userID {
			respondEphemeral(s, i, "這個使用者是 config.json 的固定管理員，不能用這個指令移除。")
			return
		}
	}

	if !adminState.AdminUsers[userID] {
		respondEphemeral(s, i, "這個使用者目前不是動態管理員。")
		return
	}

	delete(adminState.AdminUsers, userID)
	saveAdminState()

	respondEphemeral(s, i, fmt.Sprintf(
		"已移除 %s 的管理員權限。",
		formatPlayerLabel(userID),
	))
}

func handleAdminTesterGrantCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options

	playerOption := findOption(options, "player")
	if playerOption == nil {
		respondEphemeral(s, i, "缺少必要參數。")
		return
	}

	userID, _, exists, err := resolveRegisteredPlayer(playerOption.StringValue())
	if err != nil {
		respondEphemeral(s, i, err.Error())
		return
	}
	if !exists {
		respondEphemeral(s, i, "這個玩家還沒有註冊資料，請先建立名單。")
		return
	}

	adminState.TesterUsers[userID] = true
	saveAdminState()

	respondEphemeral(s, i, fmt.Sprintf(
		"已將 %s 加入測試員名單。",
		formatPlayerLabel(userID),
	))
}

func handleAdminTesterRevokeCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options

	playerOption := findOption(options, "player")
	if playerOption == nil {
		respondEphemeral(s, i, "缺少必要參數。")
		return
	}

	userID, _, exists, err := resolveRegisteredPlayer(playerOption.StringValue())
	if err != nil {
		respondEphemeral(s, i, err.Error())
		return
	}
	if !exists {
		respondEphemeral(s, i, "這個玩家目前不在名單內。")
		return
	}

	if !adminState.TesterUsers[userID] {
		respondEphemeral(s, i, "這個使用者目前不是測試員。")
		return
	}

	delete(adminState.TesterUsers, userID)
	saveAdminState()

	respondEphemeral(s, i, fmt.Sprintf(
		"已移除 %s 的測試員權限。",
		formatPlayerLabel(userID),
	))
}

func handleAdminSignupCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options

	playerOption := findOption(options, "player")
	dayOption := findOption(options, "day")
	if playerOption == nil || dayOption == nil {
		respondEphemeral(s, i, "缺少必要參數。")
		return
	}

	userID, _, exists, err := resolveRegisteredPlayer(playerOption.StringValue())
	if err != nil {
		respondEphemeral(s, i, err.Error())
		return
	}
	if !exists {
		respondEphemeral(s, i, "這個玩家還沒有資料，請先使用 /a_addplayer 建立名單。")
		return
	}
	day := dayOption.StringValue()
	weekKey := getManagedSignupWeekKey()

	if isSignupDayFull(weeklySignups, weekKey, userID, day) {
		respondEphemeral(s, i, fmt.Sprintf(
			"%s 已額滿 %d 人，無法再手動報名。",
			getDayLabel(day),
			maxSignupUsersPerDay,
		))
		return
	}

	addUserSignupDay(weekKey, userID, day)

	respondEphemeral(s, i, fmt.Sprintf(
		"已幫 %s 手動報名 %s。",
		formatPlayerLabel(userID),
		getDayLabel(day),
	))
}

func handleAdminUnsignupCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options

	playerOption := findOption(options, "player")
	dayOption := findOption(options, "day")
	if playerOption == nil || dayOption == nil {
		respondEphemeral(s, i, "缺少必要參數。")
		return
	}

	userID, _, exists, err := resolveRegisteredPlayer(playerOption.StringValue())
	if err != nil {
		respondEphemeral(s, i, err.Error())
		return
	}
	if !exists {
		respondEphemeral(s, i, "這個玩家還沒有資料，請先使用 /a_addplayer 建立名單。")
		return
	}
	day := dayOption.StringValue()

	if !removeUserSignupDay(getManagedSignupWeekKey(), userID, day) {
		respondEphemeral(s, i, fmt.Sprintf(
			"%s 本週沒有報名 %s。",
			formatPlayerLabel(userID),
			getDayLabel(day),
		))
		return
	}

	respondEphemeral(s, i, fmt.Sprintf(
		"已取消 %s 的 %s 報名。",
		formatPlayerLabel(userID),
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

func handleAdminSignupAccessCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options

	playerOption := findOption(options, "player")
	blockedOption := findOption(options, "blocked")
	if playerOption == nil || blockedOption == nil {
		respondEphemeral(s, i, "缺少必要參數。")
		return
	}

	userID, _, exists, err := resolveRegisteredPlayer(playerOption.StringValue())
	if err != nil {
		respondEphemeral(s, i, err.Error())
		return
	}
	if !exists {
		respondEphemeral(s, i, "這個玩家還沒有資料，請先使用 /a_addplayer 建立名單。")
		return
	}
	blocked := blockedOption.BoolValue()

	if blocked {
		adminState.BlockedSignupUsers[userID] = true
	} else {
		delete(adminState.BlockedSignupUsers, userID)
	}
	saveAdminState()

	respondEphemeral(s, i, fmt.Sprintf(
		"已更新 %s 的報名權限: %s",
		formatPlayerLabel(userID),
		getSignupAccessText(userID),
	))
}
