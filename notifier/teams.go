package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"log"
	"service-monitor/models"
)

func SendTeamsNotification(webhookURL, serviceName, hostIP string, port int, status string, protocol string, mentions []models.MentionUser) error {
	if webhookURL == "" {
		log.Println("⚠️ Teams webhook URL not configured, skipping notification")
		return nil
	}

	// 📌 Status Info
	emoji := "✅"
	displayStatus := "UP"
	if status == "DOWN" {
		emoji = "🔴"
		displayStatus = "DOWN"
	}

	// 🔖 Mention block
	mentionText := ""
	entities := make([]map[string]interface{}, 0, len(mentions))

	for _, user := range mentions {
		mentionText += fmt.Sprintf("<at>%s</at> ", user.Name)

		entities = append(entities, map[string]interface{}{
			"type": "mention",
			"text": fmt.Sprintf("<at>%s</at>", user.Name),
			"mentioned": map[string]interface{}{
				"id":   user.Email,
				"name": user.Name,
			},
		})
	}

	// 💬 Teams card body
	cardMsg := fmt.Sprintf(
		"%s **%s** is now **%s**\n\n> **Host**: `%s`\n\n> **Port**: `%d`\n\n> **Protocol**: `%s`\n\n%s",
		emoji, serviceName, displayStatus, hostIP, port, protocol, mentionText,
	)

	// 👤 Adaptive Card Payload
	cardContent := map[string]interface{}{
		"type":    "AdaptiveCard",
		"version": "1.0",
		"body": []map[string]interface{}{
			{
				"type": "TextBlock",
				"text": cardMsg,
				"wrap": true,
			},
		},
		"msteams": map[string]interface{}{
			"entities": entities,
		},
	}

	payload := map[string]interface{}{
		"attachments": []map[string]interface{}{
			{
				"contentType": "application/vnd.microsoft.card.adaptive",
				"content":     cardContent,
			},
		},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal Teams message: %w", err)
	}

	req, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("failed to create Teams request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to post to Teams: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("Teams webhook returned error status: %s", resp.Status)
	}

	return nil
}
