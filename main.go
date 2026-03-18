package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

var appConfig Config

func main() {
	if err := ensureProjectFiles(); err != nil {
		log.Fatal(err)
	}

	cfg := loadConfig()
	appConfig = cfg

	loadSignups()
	loadTestSignups()
	loadProfiles()
	loadAdminState()
	loadSignupScheduleState()

	dg, err := discordgo.New("Bot " + cfg.Token)
	if err != nil {
		log.Fatal("建立 Discord session 失敗:", err)
	}

	dg.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsMessageContent

	setupBotHandlers(dg)

	err = dg.Open()
	if err != nil {
		log.Fatal("連線 Discord 失敗:", err)
	}

	err = registerBaseCommands(dg, cfg)
	if err != nil {
		log.Fatal("註冊基礎 slash command 失敗:", err)
	}

	err = registerAdminCommands(dg, cfg)
	if err != nil {
		log.Fatal("註冊 admin slash command 失敗:", err)
	}

	fmt.Println("Discord Bot 已啟動，按 Ctrl+C 結束")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	dg.Close()
}

func registerBaseCommands(dg *discordgo.Session, cfg Config) error {
	commands := []*discordgo.ApplicationCommand{
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
