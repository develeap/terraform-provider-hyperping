// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"strings"
	"testing"
)

// TestValidateLocalizedText tests all language paths to increase coverage from 53.8% to 100%
func TestValidateLocalizedText(t *testing.T) {
	t.Run("empty localized text is valid", func(t *testing.T) {
		text := LocalizedText{}
		err := validateLocalizedText("field", text, 100)
		if err != nil {
			t.Errorf("expected no error for empty text, got: %v", err)
		}
	})

	t.Run("english text within limit", func(t *testing.T) {
		text := LocalizedText{En: "Hello"}
		err := validateLocalizedText("field", text, 100)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("english text exceeds limit", func(t *testing.T) {
		text := LocalizedText{En: strings.Repeat("a", 101)}
		err := validateLocalizedText("field", text, 100)
		if err == nil {
			t.Error("expected error for text exceeding max length")
		}
		if err != nil && !strings.Contains(err.Error(), "field.en") {
			t.Errorf("expected error to mention 'field.en', got: %v", err)
		}
	})

	t.Run("french text within limit", func(t *testing.T) {
		text := LocalizedText{Fr: "Bonjour"}
		err := validateLocalizedText("field", text, 100)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("french text exceeds limit", func(t *testing.T) {
		text := LocalizedText{Fr: strings.Repeat("b", 101)}
		err := validateLocalizedText("field", text, 100)
		if err == nil {
			t.Error("expected error for French text exceeding max length")
		}
		if err != nil && !strings.Contains(err.Error(), "field.fr") {
			t.Errorf("expected error to mention 'field.fr', got: %v", err)
		}
	})

	t.Run("german text within limit", func(t *testing.T) {
		text := LocalizedText{De: "Guten Tag"}
		err := validateLocalizedText("field", text, 100)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("german text exceeds limit", func(t *testing.T) {
		text := LocalizedText{De: strings.Repeat("c", 101)}
		err := validateLocalizedText("field", text, 100)
		if err == nil {
			t.Error("expected error for German text exceeding max length")
		}
		if err != nil && !strings.Contains(err.Error(), "field.de") {
			t.Errorf("expected error to mention 'field.de', got: %v", err)
		}
	})

	t.Run("spanish text within limit", func(t *testing.T) {
		text := LocalizedText{Es: "Hola"}
		err := validateLocalizedText("field", text, 100)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("spanish text exceeds limit", func(t *testing.T) {
		text := LocalizedText{Es: strings.Repeat("d", 101)}
		err := validateLocalizedText("field", text, 100)
		if err == nil {
			t.Error("expected error for Spanish text exceeding max length")
		}
		if err != nil && !strings.Contains(err.Error(), "field.es") {
			t.Errorf("expected error to mention 'field.es', got: %v", err)
		}
	})

	t.Run("multiple languages all within limit", func(t *testing.T) {
		text := LocalizedText{
			En: "Hello",
			Fr: "Bonjour",
			De: "Hallo",
			Es: "Hola",
		}
		err := validateLocalizedText("field", text, 100)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("multiple languages with one exceeding limit", func(t *testing.T) {
		text := LocalizedText{
			En: "Hello",
			Fr: strings.Repeat("x", 101), // Exceeds
			De: "Hallo",
			Es: "Hola",
		}
		err := validateLocalizedText("field", text, 100)
		if err == nil {
			t.Error("expected error when one language exceeds limit")
		}
		if err != nil && !strings.Contains(err.Error(), "field.fr") {
			t.Errorf("expected error to mention 'field.fr', got: %v", err)
		}
	})

	t.Run("all languages at exact limit", func(t *testing.T) {
		text := LocalizedText{
			En: strings.Repeat("a", 50),
			Fr: strings.Repeat("b", 50),
			De: strings.Repeat("c", 50),
			Es: strings.Repeat("d", 50),
		}
		err := validateLocalizedText("field", text, 50)
		if err != nil {
			t.Errorf("expected no error for text at exact limit, got: %v", err)
		}
	})

	t.Run("validation with zero max length", func(t *testing.T) {
		text := LocalizedText{En: "a"}
		err := validateLocalizedText("field", text, 0)
		if err == nil {
			t.Error("expected error for non-empty text with zero max length")
		}
	})

	t.Run("empty strings in all fields", func(t *testing.T) {
		// Empty strings should skip validation
		text := LocalizedText{
			En: "",
			Fr: "",
			De: "",
			Es: "",
		}
		err := validateLocalizedText("field", text, 5)
		if err != nil {
			t.Errorf("expected no error for empty strings, got: %v", err)
		}
	})

	t.Run("partial population with varying lengths", func(t *testing.T) {
		text := LocalizedText{
			En: "Short",
			De: strings.Repeat("x", 50),
			// Fr and Es empty
		}
		err := validateLocalizedText("description", text, 100)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})
}
