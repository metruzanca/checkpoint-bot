package commands

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/charmbracelet/log"
	"github.com/metruzanca/checkpoint-bot/internal/database"
	"github.com/metruzanca/checkpoint-bot/internal/database/queries"
	"github.com/metruzanca/checkpoint-bot/internal/util"
)

var CreateCheckpointCmd = &Command{
	ApplicationCommand: discordgo.ApplicationCommand{
		Name:        "checkpoint",
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

		// Ensure guild exists in database
		_, err = db.GetGuild(context.Background(), i.GuildID)
		if err == sql.ErrNoRows {
			// Guild doesn't exist, create it
			// Get guild info from Discord to get owner ID
			guild, err := s.Guild(i.GuildID)
			if err != nil {
				log.Error("cannot get guild from Discord", "err", err, "guild", i.GuildID)
				ErrorResponse(s, i, "error getting guild information")
				return
			}

			// Create guild with default timezone (UTC for now)
			_, err = db.CreateGuild(context.Background(), queries.CreateGuildParams{
				GuildID:  i.GuildID,
				Timezone: "UTC",
				OwnerID:  guild.OwnerID,
			})
			if err != nil {
				log.Error("cannot create guild", "err", err, "guild", i.GuildID)
				ErrorResponse(s, i, "error creating guild")
				return
			}
		} else if err != nil {
			// Some other error occurred
			log.Error("cannot get guild", "err", err, "guild", i.GuildID)
			ErrorResponse(s, i, "error checking guild")
			return
		}

		scheduledAtStr := scheduledAt.Format(time.RFC3339)

		// Check if there's already an upcoming checkpoint for this guild+channel
		existingUpcomingCheckpoint, err := db.GetUpcomingCheckpointByGuildAndChannel(context.Background(), queries.GetUpcomingCheckpointByGuildAndChannelParams{
			GuildID:   i.GuildID,
			ChannelID: i.ChannelID,
		})
		if err == nil {
			// An upcoming checkpoint already exists for this guild+channel
			log.Info("upcoming checkpoint already exists for guild+channel", "channel", i.ChannelID, "guild", i.GuildID, "user", i.Member.User.ID, "existing_checkpoint_id", existingUpcomingCheckpoint.ID)

			// Create embed using the same format as get-checkpoints
			embed := createCheckpointEmbed(*existingUpcomingCheckpoint)

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "An upcoming checkpoint already exists for this channel:",
					Embeds:  []*discordgo.MessageEmbed{embed},
				},
			})
			return
		} else if err != sql.ErrNoRows {
			// Some other error occurred
			log.Error("cannot check for existing upcoming checkpoint", "err", err, "channel", i.ChannelID, "guild", i.GuildID)
			ErrorResponse(s, i, "error checking for existing checkpoint")
			return
		}

		// No upcoming checkpoint exists, proceed with creation
		checkpoint, err := db.CreateCheckpoint(context.Background(), queries.CreateCheckpointParams{
			ScheduledAt: scheduledAtStr,
			ChannelID:   i.ChannelID,
			GuildID:     i.GuildID,
			DiscordUser: i.Member.User.ID,
		})

		if err != nil {
			// Check if it's a unique constraint violation (exact same time)
			errStr := err.Error()
			if strings.Contains(errStr, "UNIQUE constraint failed") || strings.Contains(errStr, "constraint failed") {
				// Duplicate checkpoint exists, get it and show it
				existingCheckpoint, err := db.GetCheckpointByScheduledAtAndChannel(context.Background(), queries.GetCheckpointByScheduledAtAndChannelParams{
					ScheduledAt: scheduledAtStr,
					ChannelID:   i.ChannelID,
				})
				if err != nil {
					log.Error("cannot get existing checkpoint", "err", err, "channel", i.ChannelID, "scheduled_at", scheduledAtStr)
					ErrorResponse(s, i, "A checkpoint already exists for this time, but couldn't retrieve it")
					return
				}

				log.Info("duplicate checkpoint attempted", "channel", i.ChannelID, "guild", i.GuildID, "user", i.Member.User.ID, "existing_checkpoint_id", existingCheckpoint.ID)

				// Create embed using the same format as get-checkpoints
				embed := createCheckpointEmbed(*existingCheckpoint)

				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "A checkpoint already exists for this time:",
						Embeds:  []*discordgo.MessageEmbed{embed},
					},
				})
				return
			}

			log.Error("cannot create checkpoint", "err", err, "channel", i.ChannelID, "guild", i.GuildID, "user", i.Member.User.ID)
			ErrorResponse(s, i, "error creating checkpoint")
			return
		}

		formattedDate := util.FormatCheckpointDate(scheduledAt)
		countdown := util.FormatCountdown(scheduledAt)
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{
					{
						Title:       "Checkpoint created",
						Description: fmt.Sprintf("Checkpoint #%d created for __%s__ %s", checkpoint.ID, formattedDate, countdown),
						Color:       0x0099ff,
					},
				},
			},
		})
	},
}

var ListCheckpointsCmd = &Command{
	ApplicationCommand: discordgo.ApplicationCommand{
		Name:        "get-checkpoints",
		Description: "List the upcoming checkpoints",
	},
	Handler: func(db database.CheckpointDatabase, s *discordgo.Session, i *discordgo.InteractionCreate) {
		checkpoints, err := db.GetUpcomingCheckpointsByGuildAndChannel(context.Background(), queries.GetUpcomingCheckpointsByGuildAndChannelParams{
			GuildID:   i.GuildID,
			ChannelID: i.ChannelID,
		})
		if err != nil {
			log.Error("cannot get upcoming checkpoints", "err", err, "channel", i.ChannelID, "guild", i.GuildID)
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Error getting upcoming checkpoints",
				},
			})
			return
		}

		if len(checkpoints) == 0 {
			log.Info("get-checkpoints command executed", "channel", i.ChannelID, "guild", i.GuildID, "user", i.Member.User.ID, "count", 0)
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "No upcoming checkpoints found for this channel.",
				},
			})
			return
		}

		log.Info("get-checkpoints command executed", "channel", i.ChannelID, "guild", i.GuildID, "user", i.Member.User.ID, "count", len(checkpoints))

		// Create an embed for each checkpoint
		embeds := make([]*discordgo.MessageEmbed, 0, len(checkpoints))
		for _, checkpoint := range checkpoints {
			embed, err := createCheckpointEmbedWithGoals(db, checkpoint)
			if err != nil {
				log.Error("cannot create checkpoint embed with goals", "err", err, "checkpoint_id", checkpoint.ID)
				// Fallback to embed without goals
				embed = createCheckpointEmbed(checkpoint)
			}
			embeds = append(embeds, embed)
		}

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: embeds,
			},
		})
	},
}

var PastCheckpointsCmd = &Command{
	ApplicationCommand: discordgo.ApplicationCommand{
		Name:        "past-checkpoints",
		Description: "List past checkpoints in this channel",
	},
	Handler: func(db database.CheckpointDatabase, s *discordgo.Session, i *discordgo.InteractionCreate) {
		checkpoints, err := db.GetPastCheckpointsByChannel(context.Background(), i.ChannelID)
		if err != nil {
			log.Error("cannot get past checkpoints", "err", err, "channel", i.ChannelID)
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Error getting past checkpoints",
				},
			})
			return
		}

		if len(checkpoints) == 0 {
			log.Info("past-checkpoints command executed", "channel", i.ChannelID, "guild", i.GuildID, "user", i.Member.User.ID, "count", 0)
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "No past checkpoints found in this channel.",
				},
			})
			return
		}

		log.Info("past-checkpoints command executed", "channel", i.ChannelID, "guild", i.GuildID, "user", i.Member.User.ID, "count", len(checkpoints))

		// Create an embed for each checkpoint
		embeds := make([]*discordgo.MessageEmbed, 0, len(checkpoints))
		for _, checkpoint := range checkpoints {
			embeds = append(embeds, createCheckpointEmbed(checkpoint))
		}

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: embeds,
			},
		})
	},
}

// createCheckpointEmbed creates a Discord embed for a checkpoint
func createCheckpointEmbed(checkpoint queries.Checkpoint) *discordgo.MessageEmbed {
	// Parse the scheduled_at time
	scheduledAt, err := time.Parse(time.RFC3339, checkpoint.ScheduledAt)
	var description string
	if err != nil {
		log.Error("cannot parse checkpoint scheduled_at", "err", err, "checkpoint_id", checkpoint.ID, "scheduled_at", checkpoint.ScheduledAt)
		description = fmt.Sprintf("Scheduled for %s", checkpoint.ScheduledAt)
	} else {
		formattedDate := util.FormatCheckpointDate(scheduledAt)
		countdown := util.FormatCountdown(scheduledAt)
		description = fmt.Sprintf("Scheduled for __%s__ %s", formattedDate, countdown)
	}

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("Checkpoint #%d", checkpoint.ID),
		Color:       0x0099ff,
		Description: description,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Channel",
				Value:  fmt.Sprintf("<#%s>", checkpoint.ChannelID),
				Inline: true,
			},
			{
				Name:   "Created by",
				Value:  fmt.Sprintf("<@%s>", checkpoint.DiscordUser),
				Inline: true,
			},
		},
	}

	// Add timestamp if parsing was successful
	if err == nil {
		embed.Timestamp = scheduledAt.Format(time.RFC3339)
	}

	return embed
}

// createCheckpointEmbedWithGoals creates a Discord embed for a checkpoint with goals included
func createCheckpointEmbedWithGoals(db database.CheckpointDatabase, checkpoint queries.Checkpoint) (*discordgo.MessageEmbed, error) {
	embed := createCheckpointEmbed(checkpoint)

	// Get goals for this checkpoint
	goals, err := db.GetGoalsByCheckpoint(context.Background(), checkpoint.ID)
	if err != nil {
		return embed, err
	}

	if len(goals) > 0 {
		// Build goals text with user mentions and status
		goalsText := ""
		for _, goal := range goals {
			statusEmoji := getStatusEmoji(goal.Status)
			goalsText += fmt.Sprintf("%s <@%s>:\n%s\n\n", statusEmoji, goal.DiscordUser, goal.Description)
		}

		// Truncate if total length exceeds Discord limit (1024 characters)
		if len(goalsText) > 1024 {
			// Try to fit as many complete goals as possible
			truncated := ""
			for _, goal := range goals {
				statusEmoji := getStatusEmoji(goal.Status)
				goalEntry := fmt.Sprintf("%s <@%s>:\n%s\n\n", statusEmoji, goal.DiscordUser, goal.Description)
				if len(truncated)+len(goalEntry) > 1020 {
					truncated += "..."
					break
				}
				truncated += goalEntry
			}
			goalsText = truncated
		}

		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   fmt.Sprintf("Goals (%d)", len(goals)),
			Value:  goalsText,
			Inline: false,
		})
	}

	return embed, nil
}

// getStatusEmoji returns an emoji representation of the goal status
func getStatusEmoji(status string) string {
	switch status {
	case "completed":
		return "✅"
	case "failed":
		return "❌"
	case "incomplete":
		return "⏳"
	default:
		return "⏳"
	}
}

func init() {
	registerCommand(CreateCheckpointCmd)
	registerCommand(ListCheckpointsCmd)
	registerCommand(PastCheckpointsCmd)
}
