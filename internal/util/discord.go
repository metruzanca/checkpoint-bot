package util

import "fmt"

func GetInviteLink(botID string) string {
	return fmt.Sprintf("https://discord.com/api/oauth2/authorize?client_id=%s&permissions=8&scope=bot", botID)
}
