package commands

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/charmbracelet/log"
	"github.com/metruzanca/checkpoint-bot/internal/database"
	"github.com/metruzanca/checkpoint-bot/internal/database/queries"
)

var GoalCmd = &Command{
	ApplicationCommand: discordgo.ApplicationCommand{
		Name:        "goal",
		Description: "Set or edit your goals for the upcoming checkpoint",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionUser,
				Name:        "user",
				Description: "User whose goals to edit (admin only)",
				Required:    false,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "status",
				Description: "Set the status of your goal",
				Required:    false,
				Choices: []*discordgo.ApplicationCommandOptionChoice{
					{
						Name:  "incomplete",
						Value: "incomplete",
					},
					{
						Name:  "failed",
						Value: "failed",
					},
					{
						Name:  "completed",
						Value: "completed",
					},
				},
			},
		},
	},
	Handler: func(db database.CheckpointDatabase, s *discordgo.Session, i *discordgo.InteractionCreate) {
		// Determine target user (self or admin override)
		targetUserID := i.Member.User.ID
		isAdminOverride := false
		var statusValue string

		options := i.ApplicationCommandData().Options
		for _, opt := range options {
			if opt.Name == "user" {
				// Check if user has admin permissions
				hasPermission := false
				if i.Member != nil {
					// Check for administrator permission
					permissions := i.Member.Permissions
					hasPermission = (permissions & discordgo.PermissionAdministrator) != 0
				}

				if !hasPermission {
					log.Warn("user attempted admin override without permission", "user", i.Member.User.ID, "guild", i.GuildID)
					ErrorResponse(s, i, "You don't have permission to edit other users' goals")
					return
				}

				targetUserID = opt.UserValue(s).ID
				isAdminOverride = true
				log.Info("admin editing user goals", "admin", i.Member.User.ID, "target_user", targetUserID, "guild", i.GuildID)
			} else if opt.Name == "status" {
				statusValue = opt.StringValue()
			}
		}

		// Get upcoming checkpoint for this guild+channel
		checkpoint, err := db.GetUpcomingCheckpointByGuildAndChannel(context.Background(), queries.GetUpcomingCheckpointByGuildAndChannelParams{
			GuildID:   i.GuildID,
			ChannelID: i.ChannelID,
		})
		if err == sql.ErrNoRows {
			log.Info("no upcoming checkpoint found", "channel", i.ChannelID, "guild", i.GuildID)
			ErrorResponse(s, i, "No upcoming checkpoint found for this channel")
			return
		} else if err != nil {
			log.Error("cannot get upcoming checkpoint", "err", err, "channel", i.ChannelID, "guild", i.GuildID)
			ErrorResponse(s, i, "Error getting upcoming checkpoint")
			return
		}

		// Check if goal already exists
		existingGoal, err := db.GetGoalByCheckpointAndUser(context.Background(), queries.GetGoalByCheckpointAndUserParams{
			CheckpointID: checkpoint.ID,
			DiscordUser:  targetUserID,
		})

		// If status is provided and goal exists, update status immediately
		if statusValue != "" && err == nil {
			err = db.UpdateGoalStatus(context.Background(), queries.UpdateGoalStatusParams{
				Status:       statusValue,
				CheckpointID: checkpoint.ID,
				DiscordUser:  targetUserID,
			})
			if err != nil {
				log.Error("cannot update goal status", "err", err, "checkpoint_id", checkpoint.ID, "user", targetUserID, "status", statusValue)
				ErrorResponse(s, i, "Error updating goal status")
				return
			}
			log.Info("goal status updated", "checkpoint_id", checkpoint.ID, "user", targetUserID, "status", statusValue)
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("Goal status updated to %s!", statusValue),
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			return
		}

		// If status is provided but goal doesn't exist, inform user they need to create goal first
		if statusValue != "" && err == sql.ErrNoRows {
			ErrorResponse(s, i, "You must create a goal first before setting its status. Use /goal without the status parameter to create one.")
			return
		}

		if err != nil && err != sql.ErrNoRows {
			log.Error("cannot get existing goal", "err", err, "checkpoint_id", checkpoint.ID, "user", targetUserID)
			ErrorResponse(s, i, "Error checking for existing goal")
			return
		}

		var goalText string
		if err == nil {
			// Goal exists, pre-fill with existing text
			goalText = existingGoal.Description
		}

		// Create modal
		modalTitle := "Set Goals"
		if isAdminOverride {
			modalTitle = fmt.Sprintf("Edit Goals for User")
		}

		// Include status in custom ID if provided (for setting status after goal creation)
		customID := fmt.Sprintf("goal_modal_%d_%s", checkpoint.ID, targetUserID)
		if statusValue != "" {
			customID = fmt.Sprintf("goal_modal_%d_%s_%s", checkpoint.ID, targetUserID, statusValue)
		}

		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseModal,
			Data: &discordgo.InteractionResponseData{
				CustomID: customID,
				Title:    modalTitle,
				Components: []discordgo.MessageComponent{
					discordgo.ActionsRow{
						Components: []discordgo.MessageComponent{
							discordgo.TextInput{
								CustomID:    "goal_text",
								Label:       "Goals",
								Style:       discordgo.TextInputParagraph,
								Placeholder: "Enter your goals for this checkpoint...",
								Value:       goalText,
								Required:    true,
								MaxLength:   2000,
								MinLength:   1,
							},
						},
					},
				},
			},
		})

		if err != nil {
			log.Error("cannot respond with modal", "err", err, "channel", i.ChannelID, "guild", i.GuildID)
			ErrorResponse(s, i, "Error opening goal editor")
			return
		}

		log.Info("goal modal opened", "checkpoint_id", checkpoint.ID, "user", targetUserID, "is_admin_override", isAdminOverride, "has_existing_goal", err == nil, "status", statusValue)
	},
}

// HandleGoalModalSubmission handles modal submissions for goal editing
func HandleGoalModalSubmission(db database.CheckpointDatabase, s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.ModalSubmitData()

	// Parse custom ID: goal_modal_{checkpoint_id}_{user_id} or goal_modal_{checkpoint_id}_{user_id}_{status}
	parts := strings.Split(data.CustomID, "_")
	if len(parts) < 4 || len(parts) > 5 || parts[0] != "goal" || parts[1] != "modal" {
		log.Error("invalid goal modal custom ID format", "custom_id", data.CustomID)
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Error processing goal submission",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	checkpointID, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		log.Error("cannot parse checkpoint ID from modal custom ID", "err", err, "custom_id", data.CustomID)
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Error processing goal submission",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}
	targetUserID := parts[3]
	var statusValue string
	if len(parts) == 5 {
		statusValue = parts[4]
	}

	// Get goal text from modal
	goalText := ""
	for _, component := range data.Components {
		if actionRow, ok := component.(*discordgo.ActionsRow); ok {
			for _, comp := range actionRow.Components {
				if textInput, ok := comp.(*discordgo.TextInput); ok && textInput.CustomID == "goal_text" {
					goalText = textInput.Value
					break
				}
			}
		}
	}

	if goalText == "" {
		log.Error("goal text is empty", "checkpoint_id", checkpointID, "user", targetUserID)
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Goal text cannot be empty",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	// Check if goal exists
	existingGoal, err := db.GetGoalByCheckpointAndUser(context.Background(), queries.GetGoalByCheckpointAndUserParams{
		CheckpointID: checkpointID,
		DiscordUser:  targetUserID,
	})

	if err == sql.ErrNoRows {
		// Create new goal
		_, err = db.CreateGoal(context.Background(), queries.CreateGoalParams{
			DiscordUser:  targetUserID,
			Description:  goalText,
			CheckpointID: checkpointID,
		})
		if err != nil {
			log.Error("cannot create goal", "err", err, "checkpoint_id", checkpointID, "user", targetUserID)
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Error saving goal",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			return
		}
		log.Info("goal created", "checkpoint_id", checkpointID, "user", targetUserID)

		// If status was provided, update it now
		if statusValue != "" {
			err = db.UpdateGoalStatus(context.Background(), queries.UpdateGoalStatusParams{
				Status:       statusValue,
				CheckpointID: checkpointID,
				DiscordUser:  targetUserID,
			})
			if err != nil {
				log.Error("cannot update goal status after creation", "err", err, "checkpoint_id", checkpointID, "user", targetUserID, "status", statusValue)
				// Don't fail the whole operation, just log the error
			} else {
				log.Info("goal status set after creation", "checkpoint_id", checkpointID, "user", targetUserID, "status", statusValue)
			}
		}
	} else if err != nil {
		log.Error("cannot check for existing goal", "err", err, "checkpoint_id", checkpointID, "user", targetUserID)
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Error checking for existing goal",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	} else {
		// Update existing goal
		err = db.UpdateGoalDescription(context.Background(), queries.UpdateGoalDescriptionParams{
			Description:  goalText,
			CheckpointID: checkpointID,
			DiscordUser:  targetUserID,
		})
		if err != nil {
			log.Error("cannot update goal", "err", err, "checkpoint_id", checkpointID, "user", targetUserID)
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Error updating goal",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			return
		}
		log.Info("goal updated", "checkpoint_id", checkpointID, "user", targetUserID)

		// If status was provided, update it now
		if statusValue != "" {
			err = db.UpdateGoalStatus(context.Background(), queries.UpdateGoalStatusParams{
				Status:       statusValue,
				CheckpointID: checkpointID,
				DiscordUser:  targetUserID,
			})
			if err != nil {
				log.Error("cannot update goal status after update", "err", err, "checkpoint_id", checkpointID, "user", targetUserID, "status", statusValue)
				// Don't fail the whole operation, just log the error
			} else {
				log.Info("goal status set after update", "checkpoint_id", checkpointID, "user", targetUserID, "status", statusValue)
			}
		}
	}

	// Success response
	action := "created"
	if existingGoal != nil {
		action = "updated"
	}
	statusMsg := ""
	if statusValue != "" {
		statusMsg = fmt.Sprintf(" Status set to %s.", statusValue)
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("Goal %s successfully!%s", action, statusMsg),
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

func init() {
	registerCommand(GoalCmd)
}
