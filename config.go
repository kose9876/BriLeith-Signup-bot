package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"unicode"
)

type Config struct {
	Token         string   `json:"token"`
	ApplicationID string   `json:"application_id"`
	GuildID       string   `json:"guild_id"`
	GuildIDs      []string `json:"guild_ids"`
	AdminUserIDs  []string `json:"admin_user_ids"`
}

func loadConfig() Config {
	var cfg Config
	err := readJSONFile(configFile, &cfg)
	if err != nil {
		log.Fatal("讀取 config.json 失敗:", err)
	}

	if needsInteractiveConfig(cfg) {
		cfg = completeConfigInteractively(cfg)
	}

	cfg = validateConfigOrExit(cfg)

	return cfg
}

func needsInteractiveConfig(cfg Config) bool {
	if strings.TrimSpace(cfg.Token) == "" || strings.TrimSpace(cfg.ApplicationID) == "" {
		return true
	}
	if strings.TrimSpace(cfg.GuildID) == "" && len(cfg.GuildIDs) == 0 {
		return true
	}
	if len(cfg.AdminUserIDs) == 0 {
		return true
	}
	return false
}

func completeConfigInteractively(cfg Config) Config {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("偵測到目前是首次設定或 config.json 尚未填完整，現在開始進行初始化。")
	fmt.Println("Discord Bot Token 與 Application ID 可在 Discord Developer Portal -> Applications -> 你的 Bot -> General Information / Bot 找到。")
	fmt.Println("Guild ID 與管理員 User ID 可先在 Discord 啟用開發者模式，再右鍵伺服器或使用者複製 ID。")

	cfg.Token = promptUntilValid(
		reader,
		"Discord Bot Token",
		strings.TrimSpace(cfg.Token),
		validateNonEmpty,
	)
	cfg.ApplicationID = promptUntilValid(
		reader,
		"Discord Application ID",
		strings.TrimSpace(cfg.ApplicationID),
		validateDiscordNumericID,
	)

	defaultGuildID := strings.TrimSpace(cfg.GuildID)
	if defaultGuildID == "" && len(cfg.GuildIDs) > 0 {
		defaultGuildID = strings.TrimSpace(cfg.GuildIDs[0])
	}
	cfg.GuildID = promptUntilValid(
		reader,
		"主要 Guild ID",
		defaultGuildID,
		validateDiscordNumericID,
	)
	cfg.GuildIDs = []string{cfg.GuildID}

	adminIDs := strings.Join(cfg.AdminUserIDs, ",")
	adminIDs = promptUntilValid(
		reader,
		"管理員 Discord User ID（可用逗號分隔多個）",
		adminIDs,
		validateDiscordIDCSV,
	)
	cfg.AdminUserIDs = splitAndTrimCSV(adminIDs)

	if err := writeJSONFile(configFile, cfg); err != nil {
		log.Fatal("寫入 config.json 失敗:", err)
	}

	fmt.Println("初始化完成，已更新 config.json。")
	return cfg
}

func promptUntilValid(reader *bufio.Reader, label string, defaultValue string, validate func(string) error) string {
	for {
		prompt := label
		if defaultValue != "" {
			prompt += " [" + defaultValue + "]"
		}
		prompt += ": "

		fmt.Print(prompt)
		text, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal("讀取輸入失敗:", err)
		}

		value := strings.TrimSpace(text)
		if value == "" {
			value = defaultValue
		}

		if err := validate(value); err != nil {
			fmt.Println("輸入格式錯誤：", err)
			continue
		}
		return value
	}
}

func validateConfigOrExit(cfg Config) Config {
	if err := validateNonEmpty(cfg.Token); err != nil {
		log.Fatal("config.json 缺少 token")
	}

	if err := validateDiscordNumericID(cfg.ApplicationID); err != nil {
		log.Fatal("config.json 缺少或錯誤的 application_id")
	}

	if len(cfg.GuildIDs) == 0 {
		if err := validateDiscordNumericID(cfg.GuildID); err != nil {
			log.Fatal("config.json 需要 guild_id 或 guild_ids")
		}
		cfg.GuildIDs = []string{cfg.GuildID}
	}

	if strings.TrimSpace(cfg.GuildID) == "" && len(cfg.GuildIDs) > 0 {
		cfg.GuildID = cfg.GuildIDs[0]
	}

	for _, guildID := range cfg.GuildIDs {
		if err := validateDiscordNumericID(guildID); err != nil {
			log.Fatal("config.json 的 guild_ids 含有無效 ID")
		}
	}

	if len(cfg.AdminUserIDs) == 0 {
		log.Fatal("config.json 缺少 admin_user_ids")
	}
	for _, userID := range cfg.AdminUserIDs {
		if err := validateDiscordNumericID(userID); err != nil {
			log.Fatal("config.json 的 admin_user_ids 含有無效 ID")
		}
	}

	return cfg
}

func validateNonEmpty(value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("不可為空")
	}
	return nil
}

func validateDiscordNumericID(value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return fmt.Errorf("不可為空")
	}
	for _, r := range value {
		if !unicode.IsDigit(r) {
			return fmt.Errorf("必須是純數字 ID")
		}
	}
	return nil
}

func validateDiscordIDCSV(value string) error {
	parts := splitAndTrimCSV(value)
	if len(parts) == 0 {
		return fmt.Errorf("至少要輸入一個 ID")
	}
	for _, part := range parts {
		if err := validateDiscordNumericID(part); err != nil {
			return err
		}
	}
	return nil
}

func splitAndTrimCSV(value string) []string {
	rawParts := strings.Split(value, ",")
	parts := make([]string, 0, len(rawParts))
	for _, part := range rawParts {
		part = strings.TrimSpace(part)
		if part != "" {
			parts = append(parts, part)
		}
	}
	return parts
}
