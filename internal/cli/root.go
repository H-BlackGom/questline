package cli

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "ql",
	Short: "Questline - gamified task management",
	Long:  "A CLI tool for managing tasks with RPG-style progression",
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

// GetDBPath returns the path to the SQLite database
func GetDBPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	return filepath.Join(home, ".questline", "data.db")
}

func init() {
	// Subcommands will be added here
}

// EnsureDataDir creates the questline data directory if it doesn't exist
func EnsureDataDir() error {
	dbPath := GetDBPath()
	dir := filepath.Dir(dbPath)
	return os.MkdirAll(dir, 0755)
}
