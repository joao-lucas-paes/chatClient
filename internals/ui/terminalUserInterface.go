package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	listModel list.Model
	styles    uiStyles
	mainMenu  chatStyles
	focus			bool
	textInput string
}

type uiStyles struct {
	box      lipgloss.Style
	header   lipgloss.Style
	help     lipgloss.Style
	dialog   lipgloss.Style
	boldBlue lipgloss.Style
	color    string
}

type chatStyles struct {
	box      lipgloss.Style
	body   	 list.Model
	chat		 string
	input    lipgloss.Style
	color    string
}

type menuItem struct {
	id    string
	title string
}

func (i menuItem) Title() string       { return i.title }
func (i menuItem) Description() string { return "" }
func (i menuItem) FilterValue() string { return i.title }

func initialLeftBar() uiStyles {
	return uiStyles{
		box:      lipgloss.NewStyle().Padding(1).Border(lipgloss.NormalBorder()),
		header:   lipgloss.NewStyle().Bold(true).MarginBottom(1),
		help:     lipgloss.NewStyle().Foreground(lipgloss.Color("241")).MarginTop(1),
		dialog:   lipgloss.NewStyle().Padding(1).Border(lipgloss.RoundedBorder()).Width(50).Height(7).Align(lipgloss.Center),
		boldBlue: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("33")),
		color:    "12",
	}
}

func initialMainBar()  chatStyles {
	items := []list.Item{
		menuItem{id:"1", title:"Voce: olha, tudo bem?"},
		menuItem{id:"2", title:"Pessoa1: Nao hahahaha"},
		menuItem{id:"3", title:"Pessoa2: puts, que engraçado"},
		menuItem{id:"4", title:"Voce: olha, tudo bem?"},
		menuItem{id:"5", title:"Pessoa1: Nao hahahaha"},
		menuItem{id:"6", title:"Pessoa2: vai dormi"},
		menuItem{id:"7", title:"Voce: olha, tudo bem?"},
		menuItem{id:"8", title:"Pessoa1: Nao hahahaha"},
		menuItem{id:"9", title:"Pessoa2: dormiu????"},
	}

	delegate := list.NewDefaultDelegate()
	delegate.SetSpacing(0)
	delegate.ShowDescription = false

	list := list.New(items, delegate, 75, 10)
	list.SetShowStatusBar(false)
	list.SetFilteringEnabled(false)
	list.SetShowHelp(false)
	list.SetShowPagination(false)


	return chatStyles{
		box:	   	lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Width(75).Height(9),
		body:   	list,
		input:    lipgloss.NewStyle().
								Border(lipgloss.NormalBorder(), false, false, true, false).
								Width(73).
								Height(1),
		color:    "white",
	}
}

func InitialModel() model {
	items := []list.Item{
		menuItem{id:"1", title:"Geral"},
		menuItem{id:"2", title:"Trabalho"},
		menuItem{id:"3", title:"Amigos"},
		menuItem{id:"4", title:"Familia"},
		menuItem{id:"5", title:"Random"},
	}

	delegate := list.NewDefaultDelegate()
	delegate.SetSpacing(0)
	delegate.ShowDescription = false

	l := list.New(items, delegate, 30, 10)
	l.Title = "Conversas"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(false)

	m := model{
		listModel: l,
		styles:    initialLeftBar(),
		mainMenu:  initialMainBar(),
	}
	return m
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) View() string {
	left := m.styles.box.
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(m.styles.color)).
		Render(m.listModel.View())

	right := renderMain(m)

	return lipgloss.JoinHorizontal(lipgloss.Top, left, right)
}

func renderMain(m model) string {
	m.mainMenu.body.Title = "Chat — Geral" // depois eu coloco o nome do grupo
	body := m.mainMenu.body.View()

	toPrint := m.textInput
	if len(m.textInput) == 0 {
		toPrint = "Escreva uma mensagem..."
	}

	input := m.mainMenu.input.
						BorderForeground(lipgloss.Color("10")).
						Render(fmt.Sprintf("> %s", toPrint))

	rightInner := lipgloss.JoinVertical(lipgloss.Top, body, input)

	right := m.mainMenu.box.
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(m.mainMenu.color)).
		Render(rightInner)
	return right
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if !m.focus {
		return leftMenuUpdate(m, msg)
	} else {
		return rightMenuUpdate(m, msg)
	}
}

func leftMenuUpdate(m model, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "tab":
				return swapFocus(m)
			case "ctrl+c", "q":
				return m, tea.Quit
			default:
				var cmd tea.Cmd
				m.listModel, cmd = m.listModel.Update(msg)
				return m, cmd
			}
		default:
			return m, nil
	}
}

func rightMenuUpdate(m model, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "tab":
				return swapFocus(m)
			case "enter":
				// enviar mensagem
				return m, nil
			case "backspace":
				if len(m.textInput) > 0 {
					m.textInput = m.textInput[:len(m.textInput)-1]
				}
				return m, nil
			case "up", "down", "right", "left", "pgup", "pgdown":
				var cmd tea.Cmd
				m.mainMenu.body, cmd = m.mainMenu.body.Update(msg)
				return m, cmd
			default:
				if (msg.String() != " " || len(m.textInput) > 0) {
					m.textInput += msg.String()
				}
				return m, nil
			}
		default:
			return m, nil
	}
}

func swapFocus(m model) (tea.Model, tea.Cmd) {
	m.focus = !m.focus
	auxColor := m.mainMenu.color
	m.mainMenu.color = m.styles.color
	m.styles.color = auxColor
	return m, nil
}