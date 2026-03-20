package main

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

func registerCommands(dg *discordgo.Session, cfg Config, commands []*discordgo.ApplicationCommand) error {
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

func removeObsoleteCommands(dg *discordgo.Session, cfg Config) error {
	allowed := map[string]bool{}
	for _, command := range buildBaseCommands() {
		allowed[command.Name] = true
	}
	for _, command := range buildAdminCommands() {
		allowed[command.Name] = true
	}

	for _, guildID := range cfg.GuildIDs {
		commands, err := dg.ApplicationCommands(cfg.ApplicationID, guildID)
		if err != nil {
			return fmt.Errorf("guild %s list commands: %w", guildID, err)
		}

		for _, command := range commands {
			if allowed[command.Name] {
				continue
			}

			if err := dg.ApplicationCommandDelete(cfg.ApplicationID, guildID, command.ID); err != nil {
				return fmt.Errorf("guild %s delete command %s: %w", guildID, command.Name, err)
			}
		}
	}

	return nil
}
