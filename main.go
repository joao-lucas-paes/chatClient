package main

import (
	"chatClient/internals/ui"
	tea "github.com/charmbracelet/bubbletea"
	"fmt"
)

func main() {
	app := tea.NewProgram(ui.InitialModel(), tea.WithAltScreen())
	if _, err := app.Run(); err != nil {
		fmt.Printf("Erro: %v\n", err)
	}
}
