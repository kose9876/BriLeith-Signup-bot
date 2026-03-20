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

	err = removeObsoleteCommands(dg, cfg)
	if err != nil {
		log.Fatal("清理過期 slash command 失敗:", err)
	}

	fmt.Println("Discord Bot 已啟動，按 Ctrl+C 結束")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	dg.Close()
}
