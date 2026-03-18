package main

import (
	"log"
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

	if cfg.Token == "" {
		log.Fatal("config.json 缺少 token")
	}

	if cfg.ApplicationID == "" {
		log.Fatal("config.json 缺少 application_id")
	}

	if len(cfg.GuildIDs) == 0 {
		if cfg.GuildID == "" {
			log.Fatal("config.json 需要 guild_id 或 guild_ids")
		}
		cfg.GuildIDs = []string{cfg.GuildID}
	}

	if cfg.GuildID == "" && len(cfg.GuildIDs) > 0 {
		cfg.GuildID = cfg.GuildIDs[0]
	}

	return cfg
}
