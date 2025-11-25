// config package provides configuration loading, built on top of viper and pflag.
package config

import (
	"fmt"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type Config struct {
	Token string
}

func LoadConfig() error {
	godotenv.Load() // Load .env file

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	viper.AutomaticEnv() // Read environment variables

	pflag.Parse()
	if err := viper.BindPFlags(pflag.CommandLine); err != nil {
		return fmt.Errorf("failed to bind pflags: %w", err)
	}

	return nil
}

// Used to bind environment variables to flags in the root command
func PersistentPreRunE(cmd *cobra.Command, args []string) error {
	if err := LoadConfig(); err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	return nil
}
