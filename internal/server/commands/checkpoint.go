package commands

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/charmbracelet/log"
	"github.com/metruzanca/checkpoint-bot/internal/database"
	"github.com/metruzanca/checkpoint-bot/internal/database/queries"
	"github.com/metruzanca/checkpoint-bot/internal/util"
)

const (
	// DiscordEmbedFieldMaxLength is the maximum length for Discord embed field values
	DiscordEmbedFieldMaxLength = 1024
	// DiscordTextInputMaxLength is the maximum length for Discord text input fields
	DiscordTextInputMaxLength = 2000
)

// CreateCheckpointCmd creates a new checkpoint for a specified date and time
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
		ctx, cancel := dbContext()
		defer cancel()

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
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "date is not a valid date (expected YYYY-MM-DD)",
				},
			})
			return
		}

		// Parse time
		hour, minute, err := util.ParseTime(timeStr)
		if err != nil {
			log.Error("cannot parse time", "err", err, "time", timeStr, "channel", i.ChannelID)
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "time is not a valid time (expected HH:MM or H:MM AM/PM)",
				},
			})
			return
		}

		// Ensure guild exists in database (needed for timezone)
		guild, err := db.GetGuild(ctx, i.GuildID)
		if err == sql.ErrNoRows {
			// Guild doesn't exist, create it
			// Get guild info from Discord to get owner ID
			discordGuild, err := s.Guild(i.GuildID)
			if err != nil {
				log.Error("cannot get guild from Discord", "err", err, "guild", i.GuildID)
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "error getting guild information",
					},
				})
				return
			}

			// Create guild with default timezone (UTC for now)
			createdGuild, err := db.CreateGuild(ctx, queries.CreateGuildParams{
				GuildID:  i.GuildID,
				Timezone: "UTC",
				OwnerID:  discordGuild.OwnerID,
			})
			if err != nil {
				log.Error("cannot create guild", "err", err, "guild", i.GuildID)
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "error creating guild",
					},
				})
				return
			}
			guild = createdGuild
		} else if err != nil {
			// Some other error occurred
			log.Error("cannot get guild", "err", err, "guild", i.GuildID)
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "error checking guild",
				},
			})
			return
		}

		// Determine timezone location
		loc := time.UTC // Default to UTC
		if guild != nil && guild.Timezone != "" {
			timezoneLoc, err := time.LoadLocation(guild.Timezone)
			if err == nil {
				loc = timezoneLoc
			} else {
				log.Warn("cannot load guild timezone, using UTC", "timezone", guild.Timezone, "err", err, "guild", i.GuildID)
			}
		}

		// Combine date and time (seconds always 0) in the guild's timezone
		scheduledAt = time.Date(
			parsedDate.Year(),
			parsedDate.Month(),
			parsedDate.Day(),
			hour,
			minute,
			0,
			0,
			loc,
		)

		// Validate that checkpoint is in the future
		now := time.Now().In(loc)
		if scheduledAt.Before(now) {
			log.Warn("attempted to create checkpoint in the past", "scheduled_at", scheduledAt, "now", now, "channel", i.ChannelID, "guild", i.GuildID, "user", i.Member.User.ID)
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Cannot create checkpoint in the past. Please schedule it for a future date and time.",
				},
			})
			return
		}

		scheduledAtStr := scheduledAt.Format(time.RFC3339)

		// Check if there's already an upcoming checkpoint for this guild+channel
		existingUpcomingCheckpoint, err := db.GetUpcomingCheckpointByGuildAndChannel(ctx, queries.GetUpcomingCheckpointByGuildAndChannelParams{
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
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "error checking for existing checkpoint",
				},
			})
			return
		}

		// Check for exact duplicate (same scheduled_at and channel) before attempting insert
		// This is more efficient than relying on database constraint errors
		existingCheckpoint, err := db.GetCheckpointByScheduledAtAndChannel(ctx, queries.GetCheckpointByScheduledAtAndChannelParams{
			ScheduledAt: scheduledAtStr,
			ChannelID:   i.ChannelID,
		})
		if err == nil {
			// Exact duplicate exists
			log.Info("duplicate checkpoint attempted", "channel", i.ChannelID, "guild", i.GuildID, "user", i.Member.User.ID, "existing_checkpoint_id", existingCheckpoint.ID)
			embed := createCheckpointEmbed(*existingCheckpoint)
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "A checkpoint already exists for this time:",
					Embeds:  []*discordgo.MessageEmbed{embed},
				},
			})
			return
		} else if err != sql.ErrNoRows {
			// Error checking for duplicate
			log.Error("cannot check for duplicate checkpoint", "err", err, "channel", i.ChannelID, "scheduled_at", scheduledAtStr)
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "error checking for duplicate checkpoint",
				},
			})
			return
		}

		// No duplicate exists, proceed with creation
		checkpoint, err := db.CreateCheckpoint(ctx, queries.CreateCheckpointParams{
			ScheduledAt: scheduledAtStr,
			ChannelID:   i.ChannelID,
			GuildID:     i.GuildID,
			DiscordUser: i.Member.User.ID,
		})

		if err != nil {

			log.Error("cannot create checkpoint", "err", err, "channel", i.ChannelID, "guild", i.GuildID, "user", i.Member.User.ID)
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "error creating checkpoint",
				},
			})
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

// ListCheckpointsCmd lists all upcoming checkpoints for the current channel
var ListCheckpointsCmd = &Command{
	ApplicationCommand: discordgo.ApplicationCommand{
		Name:        "get-checkpoints",
		Description: "List the upcoming checkpoints",
	},
	Handler: func(db database.CheckpointDatabase, s *discordgo.Session, i *discordgo.InteractionCreate) {
		ctx, cancel := dbContext()
		defer cancel()

		checkpoints, err := db.GetUpcomingCheckpointsByGuildAndChannel(ctx, queries.GetUpcomingCheckpointsByGuildAndChannelParams{
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
			embed, err := createCheckpointEmbedWithGoals(ctx, db, checkpoint)
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

// PastCheckpointsCmd lists all past checkpoints for the current channel
var PastCheckpointsCmd = &Command{
	ApplicationCommand: discordgo.ApplicationCommand{
		Name:        "past-checkpoints",
		Description: "List past checkpoints in this channel",
	},
	Handler: func(db database.CheckpointDatabase, s *discordgo.Session, i *discordgo.InteractionCreate) {
		ctx, cancel := dbContext()
		defer cancel()

		checkpoints, err := db.GetPastCheckpointsByChannel(ctx, i.ChannelID)
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

// createCheckpointEmbed creates a Discord embed for a checkpoint with formatted date and countdown
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

// createCheckpointEmbedWithGoals creates a Discord embed for a checkpoint including associated goals
// Goals are truncated if they exceed Discord's embed field length limit
func createCheckpointEmbedWithGoals(ctx context.Context, db database.CheckpointDatabase, checkpoint queries.Checkpoint) (*discordgo.MessageEmbed, error) {
	embed := createCheckpointEmbed(checkpoint)

	// Get goals for this checkpoint
	goals, err := db.GetGoalsByCheckpoint(ctx, checkpoint.ID)
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

		// Truncate if total length exceeds Discord limit
		if len(goalsText) > DiscordEmbedFieldMaxLength {
			// Try to fit as many complete goals as possible
			truncated := ""
			for _, goal := range goals {
				statusEmoji := getStatusEmoji(goal.Status)
				goalEntry := fmt.Sprintf("%s <@%s>:\n%s\n\n", statusEmoji, goal.DiscordUser, goal.Description)
				if len(truncated)+len(goalEntry) > DiscordEmbedFieldMaxLength-4 {
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

// getStatusEmoji returns an emoji representation of a goal's status
// Returns ✅ for completed, ❌ for failed, ⏳ for incomplete or unknown
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
