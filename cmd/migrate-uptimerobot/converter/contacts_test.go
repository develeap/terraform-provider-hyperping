// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package converter

import (
	"testing"

	"github.com/develeap/terraform-provider-hyperping/cmd/migrate-uptimerobot/uptimerobot"
)

func TestCategorizeAlertContacts(t *testing.T) {
	contacts := []uptimerobot.AlertContact{
		{ID: "1", Type: 2, Value: "a@example.com", FriendlyName: "Alice"},
		{ID: "2", Type: 3, Value: "+15555555555", FriendlyName: "Phone"},
		{ID: "3", Type: 4, Value: "https://hook.example.com", FriendlyName: "Hook"},
		{ID: "4", Type: 11, FriendlyName: "Slack #ops"},
		{ID: "5", Type: 14, FriendlyName: "PagerDuty Primary"},
		{ID: "6", Type: 99, FriendlyName: "Some Other"},
	}

	got := CategorizeAlertContacts(contacts)

	if len(got.Emails) != 1 || got.Emails[0] != "a@example.com" {
		t.Errorf("Emails = %v, want [a@example.com]", got.Emails)
	}
	if len(got.SMSPhones) != 1 || got.SMSPhones[0] != "+15555555555" {
		t.Errorf("SMSPhones = %v", got.SMSPhones)
	}
	if len(got.Webhooks) != 1 || got.Webhooks[0] != "https://hook.example.com" {
		t.Errorf("Webhooks = %v", got.Webhooks)
	}
	if len(got.Slack) != 1 || got.Slack[0] != "Slack #ops" {
		t.Errorf("Slack = %v", got.Slack)
	}
	if len(got.PagerDuty) != 1 || got.PagerDuty[0] != "PagerDuty Primary" {
		t.Errorf("PagerDuty = %v", got.PagerDuty)
	}
	if len(got.Other) != 1 || got.Other[0] != "Some Other" {
		t.Errorf("Other = %v", got.Other)
	}
}

func TestCategorizeAlertContacts_Empty(t *testing.T) {
	got := CategorizeAlertContacts(nil)
	if got == nil {
		t.Fatal("got nil, want non-nil empty struct")
	}
	if len(got.Emails) != 0 || len(got.SMSPhones) != 0 || len(got.Webhooks) != 0 ||
		len(got.Slack) != 0 || len(got.PagerDuty) != 0 || len(got.Other) != 0 {
		t.Errorf("expected all empty slices, got %+v", got)
	}
}

func TestGetAlertContactTypeName(t *testing.T) {
	// Slice rather than map for deterministic subtest order under -shuffle.
	cases := []struct {
		typeID int
		want   string
	}{
		{2, "Email"},
		{3, "SMS"},
		{4, "Webhook"},
		{5, "Twitter"},
		{6, "Boxcar"},
		{7, "Web-Hook"},
		{8, "Pushbullet"},
		{9, "Zapier"},
		{10, "Pushover"},
		{11, "Slack"},
		{12, "HipChat"},
		{13, "Google Hangouts Chat"},
		{14, "PagerDuty"},
		{15, "Opsgenie"},
		{16, "VictorOps"},
		{17, "Microsoft Teams"},
		{0, "Unknown"},
		{999, "Unknown"},
	}
	for _, tc := range cases {
		if got := GetAlertContactTypeName(tc.typeID); got != tc.want {
			t.Errorf("GetAlertContactTypeName(%d) = %q, want %q", tc.typeID, got, tc.want)
		}
	}
}
