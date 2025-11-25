package config

import (
	"os"
	"testing"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLoadConfig tests the LoadConfig function for basic functionality
// and source priority.
func TestLoadConfig(t *testing.T) {
	// Reset viper and pflag for a clean test environment
	viper.Reset()
	pflag.CommandLine = pflag.NewFlagSet(os.Args[0], pflag.ExitOnError)

	// 1. Set Defaults
	viper.SetDefault("testkey", "default_value")
	viper.SetDefault("anotherkey", 10)

	// 2. Define a flag
	pflag.String("testkey", "", "A test key for flags")
	pflag.Parse() // Parse flags to make them available for binding

	// Load config
	err := LoadConfig()
	require.NoError(t, err, "LoadConfig failed")

	// Test default value
	assert.Equal(t, "default_value", viper.GetString("testkey"), "testkey should be 'default_value'")
	assert.Equal(t, 10, viper.GetInt("anotherkey"), "anotherkey should be 10")
}

// TestLoadConfig_EnvironmentVariables tests that environment variables
// correctly override defaults.
func TestLoadConfig_EnvironmentVariables(t *testing.T) {
	// Reset viper and pflag for a clean test environment
	viper.Reset()
	pflag.CommandLine = pflag.NewFlagSet(os.Args[0], pflag.ExitOnError)

	// Set a default
	viper.SetDefault("envtestkey", "default_env")

	// Set an environment variable (normalized)
	os.Setenv("ENVTESTKEY", "env_value")
	defer os.Unsetenv("ENVTESTKEY") // Clean up env var after test

	// Define a flag, but don't set it via command line
	pflag.String("envtestkey", "", "A test key for flags")
	pflag.Parse()

	// Load config
	err := LoadConfig()
	require.NoError(t, err, "LoadConfig failed")

	// Environment variable should override default
	assert.Equal(t, "env_value", viper.GetString("envtestkey"), "envtestkey should be 'env_value' from environment")
}

// TestLoadConfig_Flags tests that command-line flags correctly override
// environment variables and defaults.
func TestLoadConfig_Flags(t *testing.T) {
	// Reset viper and pflag for a clean test environment
	viper.Reset()
	pflag.CommandLine = pflag.NewFlagSet(os.Args[0], pflag.ExitOnError)

	// Set a default
	viper.SetDefault("flagtestkey", "default_flag")

	// Set an environment variable (which should be overridden)
	os.Setenv("FLAGTESTKEY", "env_flag")
	defer os.Unsetenv("FLAGTESTKEY")

	// Define and set a flag
	pflag.String("flagtestkey", "flag_value", "A test key for flags")
	// Manually set a flag value as if from command line parsing
	err := pflag.CommandLine.Set("flagtestkey", "flag_value")
	require.NoError(t, err, "Failed to set pflag")
	pflag.Parse()

	// Load config
	err = LoadConfig()
	require.NoError(t, err, "LoadConfig failed")

	// Flag should override environment variable and default
	assert.Equal(t, "flag_value", viper.GetString("flagtestkey"), "flagtestkey should be 'flag_value' from flag")
}
