package cmd

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/GrapeInTheTree/chiliz-cli/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
)

// runTUI launches the interactive TUI mode (same behavior as original main.go)
func runTUI() error {
	// Setup logging to file (not stderr, to avoid TUI interference)
	logFile, err := os.OpenFile("chiliz.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	defer logFile.Close()

	logger := slog.New(slog.NewJSONHandler(logFile, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Initialize and run TUI
	initialModel := tui.NewModel()
	p := tea.NewProgram(initialModel, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("TUI error: %w", err)
	}
	return nil
}
