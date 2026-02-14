// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

// Package interactive provides interactive CLI utilities for migration tools.
package interactive

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
)

// Config holds interactive mode configuration.
type Config struct {
	Stdin  terminal.FileReader
	Stdout terminal.FileWriter
	Stderr terminal.FileWriter
}

// DefaultConfig returns default configuration using standard I/O.
func DefaultConfig() *Config {
	return &Config{
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
}

// APIKeyValidator validates API key format.
func APIKeyValidator(val interface{}) error {
	str, ok := val.(string)
	if !ok {
		return errors.New("invalid type")
	}

	if str == "" {
		return errors.New("api key is required")
	}

	return nil
}

// SourceAPIKeyValidator validates source platform API keys.
func SourceAPIKeyValidator(platform string) survey.Validator {
	return func(val interface{}) error {
		if err := APIKeyValidator(val); err != nil {
			return err
		}

		str, ok := val.(string)
		if !ok {
			return errors.New("invalid type")
		}

		switch platform {
		case "betterstack":
			// Better Stack tokens can vary in format
			if len(str) < 10 {
				return errors.New("API key seems too short")
			}
		case "uptimerobot":
			// UptimeRobot keys typically start with 'u' or 'm'
			if !strings.HasPrefix(str, "u") && !strings.HasPrefix(str, "m") {
				return errors.New("UptimeRobot API keys typically start with 'u' or 'm'")
			}
		case "pingdom":
			// Pingdom tokens are typically longer strings
			if len(str) < 20 {
				return errors.New("pingdom API key seems too short")
			}
		}

		return nil
	}
}

// HyperpingAPIKeyValidator validates Hyperping API key format.
func HyperpingAPIKeyValidator(val interface{}) error {
	if err := APIKeyValidator(val); err != nil {
		return err
	}

	str, ok := val.(string)
	if !ok {
		return errors.New("invalid type")
	}
	if !strings.HasPrefix(str, "sk_") {
		return errors.New("hyperping API keys must start with 'sk_'")
	}

	return nil
}

// FilePathValidator validates file path input.
func FilePathValidator(val interface{}) error {
	str, ok := val.(string)
	if !ok {
		return errors.New("invalid type")
	}

	if str == "" {
		return errors.New("file path is required")
	}

	// Check for invalid characters
	if strings.ContainsAny(str, "\x00") {
		return errors.New("file path contains invalid characters")
	}

	return nil
}

// Prompter handles interactive prompts.
type Prompter struct {
	config *Config
}

// NewPrompter creates a new prompter.
func NewPrompter(config *Config) *Prompter {
	if config == nil {
		config = DefaultConfig()
	}
	return &Prompter{config: config}
}

// AskString prompts for a string input.
func (p *Prompter) AskString(message, defaultValue, help string, validator survey.Validator) (string, error) {
	var result string
	prompt := &survey.Input{
		Message: message,
		Default: defaultValue,
		Help:    help,
	}

	opts := []survey.AskOpt{}
	if validator != nil {
		opts = append(opts, survey.WithValidator(validator))
	}

	opts = append(opts, survey.WithStdio(p.config.Stdin, p.config.Stdout, p.config.Stderr))

	err := survey.AskOne(prompt, &result, opts...)
	if err != nil {
		if err == terminal.InterruptErr {
			return "", errors.New("operation cancelled by user")
		}
		return "", err
	}

	return result, nil
}

// AskPassword prompts for password/secret input.
func (p *Prompter) AskPassword(message, help string, validator survey.Validator) (string, error) {
	var result string
	prompt := &survey.Password{
		Message: message,
		Help:    help,
	}

	opts := []survey.AskOpt{}
	if validator != nil {
		opts = append(opts, survey.WithValidator(validator))
	}

	opts = append(opts, survey.WithStdio(p.config.Stdin, p.config.Stdout, p.config.Stderr))

	err := survey.AskOne(prompt, &result, opts...)
	if err != nil {
		if err == terminal.InterruptErr {
			return "", errors.New("operation cancelled by user")
		}
		return "", err
	}

	return result, nil
}

// AskConfirm prompts for yes/no confirmation.
func (p *Prompter) AskConfirm(message string, defaultValue bool) (bool, error) {
	var result bool
	prompt := &survey.Confirm{
		Message: message,
		Default: defaultValue,
	}

	err := survey.AskOne(
		prompt,
		&result,
		survey.WithStdio(p.config.Stdin, p.config.Stdout, p.config.Stderr),
	)
	if err != nil {
		if err == terminal.InterruptErr {
			return false, errors.New("operation cancelled by user")
		}
		return false, err
	}

	return result, nil
}

// AskSelect prompts for selection from a list.
func (p *Prompter) AskSelect(message string, options []string, defaultValue string) (string, error) {
	var result string
	prompt := &survey.Select{
		Message: message,
		Options: options,
		Default: defaultValue,
	}

	err := survey.AskOne(
		prompt,
		&result,
		survey.WithStdio(p.config.Stdin, p.config.Stdout, p.config.Stderr),
	)
	if err != nil {
		if err == terminal.InterruptErr {
			return "", errors.New("operation cancelled by user")
		}
		return "", err
	}

	return result, nil
}

// PrintHeader prints a section header.
func (p *Prompter) PrintHeader(title string) {
	fmt.Fprintf(p.config.Stderr, "\n%s\n", title)
}

// PrintSuccess prints a success message.
func (p *Prompter) PrintSuccess(message string) {
	fmt.Fprintf(p.config.Stderr, "✅ %s\n", message)
}

// PrintError prints an error message.
func (p *Prompter) PrintError(message string) {
	fmt.Fprintf(p.config.Stderr, "❌ %s\n", message)
}

// PrintWarning prints a warning message.
func (p *Prompter) PrintWarning(message string) {
	fmt.Fprintf(p.config.Stderr, "⚠️  %s\n", message)
}

// PrintInfo prints an info message.
func (p *Prompter) PrintInfo(message string) {
	fmt.Fprintf(p.config.Stderr, "ℹ️  %s\n", message)
}
