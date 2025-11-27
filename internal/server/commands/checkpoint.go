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
	"github.com/metruzanca/checkpoint-bot/internal/util"
)

var CreateCheckpointCmd = &Command{
	ApplicationCommand: discordgo.ApplicationCommand{
		Name:        "create-checkpoint",
		Description: "Create a new checkpoint",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "date",
				Description: "The date of the checkpoint (YYYY-MM-DD)",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "time",
				Description: "The time of the checkpoint (HH:MM or H:MM AM/PM)",
				Required:    true,
			},
		},
	},
	Handler: func(db database.CheckpointDatabase, s *discordgo.Session, i *discordgo.InteractionCreate) {
		var scheduledAt time.Time
		options := i.ApplicationCommandData().Options

		// Find date and time options
		var dateStr, timeStr string
		for _, opt := range options {
			if opt.Name == "date" {
				dateStr = opt.StringValue()
			} else if opt.Name == "time" {
				timeStr = opt.StringValue()
			}
		}

		// Parse date (YYYY-MM-DD)
		parsedDate, err := util.ParseDate(dateStr)
		if err != nil {
			log.Error("cannot parse date", "channel", i.ChannelID, "date", dateStr, "time", timeStr)
			ErrorResponse(s, i, "date is not a valid date (expected YYYY-MM-DD)")
			return
		}

		// Parse time
		hour, minute, err := util.ParseTime(timeStr)
		if err != nil {
			log.Error("cannot parse time", "err", err, "time", timeStr, "channel", i.ChannelID)
			ErrorResponse(s, i, "time is not a valid time (expected HH:MM or H:MM AM/PM)")
			return
		}

		// Combine date and time (seconds always 0)
		scheduledAt = time.Date(
			parsedDate.Year(),
			parsedDate.Month(),
			parsedDate.Day(),
			hour,
			minute,
			0,
			0,
			time.Local,
		)

		checkpoint, err := db.CreateCheckpoint(context.Background(), queries.CreateCheckpointParams{
			ScheduledAt: scheduledAt.Format(time.RFC3339),
			ChannelID:   i.ChannelID,
		})

		if err != nil {
			ErrorResponse(s, i, "error creating checkpoint")
			return
		}

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{
					{
						Title:       "Checkpoint created",
						Description: fmt.Sprintf("Checkpoint #%s created for %s", checkpoint.ID, util.FormatNaturalDate(scheduledAt)),
						Color:       0x0099ff,
					},
				},
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
			ErrorResponse(s, i, "Error fetching checkpoint")
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
