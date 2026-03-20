package main

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

func logInteractionCommand(i *discordgo.InteractionCreate) {
	userName, userID := getInteractionUser(i)
	fmt.Printf(
		"[interaction] command | user=%s (%s) | guild=%s | channel=%s | action=%s\n",
		userName,
		userID,
		emptyLogValue(i.GuildID),
		emptyLogValue(i.ChannelID),
		i.ApplicationCommandData().Name,
	)
}

func logInteractionComponent(i *discordgo.InteractionCreate) {
	userName, userID := getInteractionUser(i)
	fmt.Printf(
		"[interaction] component | user=%s (%s) | guild=%s | channel=%s | action=%s\n",
		userName,
		userID,
		emptyLogValue(i.GuildID),
		emptyLogValue(i.ChannelID),
		i.MessageComponentData().CustomID,
	)
}

func getInteractionUser(i *discordgo.InteractionCreate) (string, string) {
	if i.Member != nil && i.Member.User != nil {
		return i.Member.User.Username, i.Member.User.ID
	}
	if i.User != nil {
		return i.User.Username, i.User.ID
	}
	return "unknown", "unknown"
}

func emptyLogValue(value string) string {
	if value == "" {
		return "-"
	}
	return value
}
