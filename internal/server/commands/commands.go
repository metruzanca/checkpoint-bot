package commands

import (
	"github.com/bwmarrin/discordgo"
)

var CreateCheckpoint = &discordgo.ApplicationCommand{
	Name:        "create-checkpoint",
	Description: "Create a new checkpoint",
	// Options: []*discordgo.ApplicationCommandOption{
	// 	{
	// 		Type:        discordgo.ApplicationCommandOptionString,
	// 		Name:        "description",
	// 		Description: "The description of the checkpoint",
	// 		Required:    true,
	// 	},
	// },
}
