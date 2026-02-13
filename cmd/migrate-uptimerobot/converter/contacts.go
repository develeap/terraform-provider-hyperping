// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package converter

import (
	"github.com/develeap/terraform-provider-hyperping/cmd/migrate-uptimerobot/uptimerobot"
)

// AlertContactInfo represents categorized alert contact information.
type AlertContactInfo struct {
	Emails    []string
	SMSPhones []string
	Webhooks  []string
	Slack     []string
	PagerDuty []string
	Other     []string
}

// CategorizeAlertContacts categorizes alert contacts by type.
func CategorizeAlertContacts(contacts []uptimerobot.AlertContact) *AlertContactInfo {
	info := &AlertContactInfo{
		Emails:    []string{},
		SMSPhones: []string{},
		Webhooks:  []string{},
		Slack:     []string{},
		PagerDuty: []string{},
		Other:     []string{},
	}

	for _, contact := range contacts {
		switch contact.Type {
		case 2: // Email
			info.Emails = append(info.Emails, contact.Value)
		case 3: // SMS
			info.SMSPhones = append(info.SMSPhones, contact.Value)
		case 4: // Webhook
			info.Webhooks = append(info.Webhooks, contact.Value)
		case 11: // Slack
			info.Slack = append(info.Slack, contact.FriendlyName)
		case 14: // PagerDuty
			info.PagerDuty = append(info.PagerDuty, contact.FriendlyName)
		default:
			info.Other = append(info.Other, contact.FriendlyName)
		}
	}

	return info
}

// GetAlertContactTypeName returns the human-readable name for an alert contact type.
func GetAlertContactTypeName(typeID int) string {
	typeNames := map[int]string{
		2:  "Email",
		3:  "SMS",
		4:  "Webhook",
		5:  "Twitter",
		6:  "Boxcar",
		7:  "Web-Hook",
		8:  "Pushbullet",
		9:  "Zapier",
		10: "Pushover",
		11: "Slack",
		12: "HipChat",
		13: "Google Hangouts Chat",
		14: "PagerDuty",
		15: "Opsgenie",
		16: "VictorOps",
		17: "Microsoft Teams",
	}

	if name, ok := typeNames[typeID]; ok {
		return name
	}
	return "Unknown"
}
