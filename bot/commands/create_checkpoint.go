package commands

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

var CreateCheckpointCommand = &discordgo.ApplicationCommand{
	Name:        "create-checkpoint",
	Description: "Create a new checkpoint",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "description",
			Description: "The description of the checkpoint",
			Required:    true,
		},
	},
}

func CreateCheckpointCallback(s *discordgo.Session, i *discordgo.InteractionCreate) {
	fmt.Println("TODO this logs twice")
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Hey there! Congratulations, you just executed your first slash command",
		},
	})
}
