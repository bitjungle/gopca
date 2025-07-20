package cmd

import (
	"testing"
)

func TestRootCommand(t *testing.T) {
	// Test that root command exists
	if rootCmd == nil {
		t.Error("rootCmd should not be nil")
	}

	// Test command attributes
	if rootCmd.Use != "complab-cli" {
		t.Errorf("Expected Use to be 'complab-cli', got '%s'", rootCmd.Use)
	}

	if rootCmd.Version != "0.1.0" {
		t.Errorf("Expected Version to be '0.1.0', got '%s'", rootCmd.Version)
	}

	// Test that subcommands are registered
	subcommands := rootCmd.Commands()
	expectedCommands := map[string]bool{
		"analyze":  false,
		"validate": false,
		"info":     false,
	}

	for _, cmd := range subcommands {
		if _, ok := expectedCommands[cmd.Use]; ok {
			expectedCommands[cmd.Use] = true
		}
	}

	for cmdName, found := range expectedCommands {
		if !found {
			t.Errorf("Expected command '%s' not found", cmdName)
		}
	}
}

func TestGlobalFlags(t *testing.T) {
	// Test verbose flag
	verboseFlag := rootCmd.PersistentFlags().Lookup("verbose")
	if verboseFlag == nil {
		t.Error("verbose flag should exist")
	}
	if verboseFlag.Shorthand != "v" {
		t.Errorf("Expected verbose shorthand to be 'v', got '%s'", verboseFlag.Shorthand)
	}

	// Test quiet flag
	quietFlag := rootCmd.PersistentFlags().Lookup("quiet")
	if quietFlag == nil {
		t.Error("quiet flag should exist")
	}
	if quietFlag.Shorthand != "q" {
		t.Errorf("Expected quiet shorthand to be 'q', got '%s'", quietFlag.Shorthand)
	}
}