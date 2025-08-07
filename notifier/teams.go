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

// SendTeamsNotification sends a notification to a Teams webhook.
// alertCount: 1 for first alert, 2 for second reminder, 3+ for third and subsequent reminders.
func SendTeamsNotification(webhookURL, serviceName, hostIP string, port int, status string, protocol string, mentions []models.MentionUser, alertCount int) error {
	if webhookURL == "" {
		log.Println("⚠️ Teams webhook URL not configured, skipping notification")
		return nil
	}

	// Determine emoji and reminder tag based on alertCount and status
	emoji := "✅"
	reminderTag := ""
	displayStatus := "UP"

	if status == "DOWN" {
		switch alertCount {
		case 1:
			emoji = "🔥"
			reminderTag = ""
		case 2:
			emoji = "🔔" // Bell emoji for second alert
			reminderTag = "**🔔 Reminder:** "
		default:
			emoji = "🚨" // Police car light for 3rd+ alert
			reminderTag = "**🚨 Final Reminder:** "
		}
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

	// 👤 Adaptive Card Payload - Modern design
	cardContent := map[string]interface{}{
		"type":    "AdaptiveCard",
		"version": "1.4",
		"$schema": "http://adaptivecards.io/schemas/adaptive-card.json",
		"body": []map[string]interface{}{
			{
				"type": "ColumnSet",
				"columns": []map[string]interface{}{
					{
						"type":  "Column",
						"width": "auto",
						"items": []map[string]interface{}{
							{
								"type":                "TextBlock",
								"text":                emoji,
								"size":                "ExtraLarge",
								"weight":              "Bolder",
								"horizontalAlignment": "Center",
								"spacing":             "None",
							},
						},
					},
					{
						"type":  "Column",
						"width": "stretch",
						"items": []map[string]interface{}{
							{
								"type":    "TextBlock",
								"text":    fmt.Sprintf("%s%s is now **%s**", reminderTag, serviceName, displayStatus),
								"weight":  "Bolder",
								"size":    "Medium",
								"wrap":    true,
								"spacing": "None",
							},
							{
								"type": "FactSet",
								"facts": []map[string]string{
									{"title": "Host:", "value": hostIP},
									{"title": "Port:", "value": fmt.Sprintf("%d", port)},
									{"title": "Protocol:", "value": protocol},
								},
								"spacing": "Small",
							},
						},
					},
				},
				"spacing": "Medium",
			},
			{
				"type":    "TextBlock",
				"text":    mentionText,
				"wrap":    true,
				"spacing": "Medium",
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
