package handlers

import (
	"context"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/charmbracelet/log"
	"github.com/metruzanca/checkpoint-bot/internal/database"
	"github.com/metruzanca/checkpoint-bot/internal/database/queries"
)

type CheckpointHandler struct {
	DiscordClient *discordgo.Session
	Database      database.CheckpointDatabase
}

func NewCheckpointHandler(discordClient *discordgo.Session, database database.CheckpointDatabase) *CheckpointHandler {
	return &CheckpointHandler{
		Database:      database,
		DiscordClient: discordClient,
	}
}

func (h *CheckpointHandler) CreateCheckpoint(s *discordgo.Session, i *discordgo.InteractionCreate) {
	checkpoint, err := h.Database.CreateCheckpoint(context.Background(), queries.CreateCheckpointParams{
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
			Content: fmt.Sprintf("Checkpoint created: %s", checkpoint.ID),
		},
	})
}
