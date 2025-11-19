package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/GrapeInTheTree/go-ethereum-butler/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Setup structured logging with slog (to file instead of stderr to not interfere with TUI)
	logFile, err := os.OpenFile("butler.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("Failed to open log file: %v\n", err)
		os.Exit(1)
	}
	defer logFile.Close()

	logger := slog.New(slog.NewJSONHandler(logFile, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Initialize TUI model
	initialModel := tui.NewModel()

	// Create and run Bubbletea program
	p := tea.NewProgram(initialModel, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
