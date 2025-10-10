package main

import (
	"chatClient/internals/ui"
	tea "github.com/charmbracelet/bubbletea"
	"fmt"
)

func init() {}

func main() {
	model 		:= ui.InitialModel()
	app 			:= tea.NewProgram(model, tea.WithAltScreen())
	model.P 	 = app

	if _, err := app.Run(); err != nil {
		fmt.Printf("Erro: %v\n", err)
	}
}
