package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/charmbracelet/log"
	"github.com/metruzanca/checkpoint-bot/internal/database"
	"github.com/metruzanca/checkpoint-bot/internal/database/queries"
)

var CreateCheckpointCmd = &Command{
	ApplicationCommand: discordgo.ApplicationCommand{
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
	},
	Handler: func(db database.CheckpointDatabase, s *discordgo.Session, i *discordgo.InteractionCreate) {
		checkpoint, err := db.CreateCheckpoint(context.Background(), queries.CreateCheckpointParams{
			ScheduledAt: time.Now().Add(time.Hour * 24 * 14).Format(time.RFC3339),
			ChannelID:   i.ChannelID,
		})

		if err != nil {
			log.Error("cannot create checkpoint", "err", err, "channel", i.ChannelID)
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Error creating checkpoint",
				},
			})
			return
		}

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Checkpoint created: %d", checkpoint.ID),
			},
		})
	},
}

var ListCheckpointsCmd = &Command{
	ApplicationCommand: discordgo.ApplicationCommand{
		Name:        "next-checkpoints",
		Description: "List the next checkpoint",
	},
	Handler: func(db database.CheckpointDatabase, s *discordgo.Session, i *discordgo.InteractionCreate) {
		checkpoint, err := db.GetUpcomingCheckpoint(context.Background())
		if err != nil {
			log.Error("cannot get upcoming checkpoint", "err", err, "channel", i.ChannelID)
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Error getting upcoming checkpoint",
				},
			})
			return
		}

		checkpointJSON, err := json.Marshal(checkpoint)
		if err != nil {
			log.Error("cannot marshal checkpoint to JSON", "err", err)
			ErrorResponse(s, i)
			return
		}

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Upcoming checkpoint:\n```json\n%s\n```", string(checkpointJSON)),
			},
		})
	},
}

func init() {
	registerCommand(CreateCheckpointCmd)
	registerCommand(ListCheckpointsCmd)
}
