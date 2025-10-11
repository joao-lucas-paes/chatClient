package ui

import (
	"bufio"
	"chatClient/internals/api"
	"fmt"
	"strconv"
	"sync"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	talkToLogin1 = "Server: Qual endereco? (caso esteja vendo novamente essa tela, algum erro na sessao pode ter acontecido)"
	talkToLogin2 = "Server: Qual porta?"
	talkToLogin3 = "Server: Qual nickname?"
	talkToLogin4 = "Server: Qual canal?"
)

type model struct {
	listModel list.Model
	styles    uiStyles
	mainMenu  chatStyles
	focus     bool
	lg        loginSystem
	textInput string
	chat      []chan api.Msg
	channels  []api.Channel
	idxChat   int
	idxMsgs		int
	msgs      []chan api.Msg
	msgsShow  []*[]api.Msg
	syncMutex []*sync.Mutex
	sendMsg   func(msg api.Msg, m model) model
	P 				*tea.Program
}

type loginSystem struct {
	isLoggin  bool
	idxLoggin int
	chat      []string
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
	box   lipgloss.Style
	body  list.Model
	input lipgloss.Style
	color string
}

type menuItem struct {
	id    string
	title string
}

func (i menuItem) Title() string       { return i.title }
func (i menuItem) Description() string { return "" }
func (i menuItem) FilterValue() string { return i.title }

func sendMsgToLoginSystem(msg api.Msg, m model) model {
	m.lg.chat[m.lg.idxLoggin] = msg.Text
	m.lg.idxLoggin++
	m.textInput = ""
	if (m.lg.isLoggin && m.lg.idxLoggin == 4) {
		m.textInput = "CLIQUE ENTER PARA PROSSEGUIR"
	}
	return m
}

func sendMsgToServerSystem(msg api.Msg, m model) model {
	if m.idxChat >= 0 && m.idxChat < len(m.msgs) {
		select {
		case m.msgs[m.idxChat] <- msg:
			msg.Text = "vc:"+msg.Text
			m.syncMutex[m.idxChat].Lock()
			*m.msgsShow[m.idxChat] = append(*m.msgsShow[m.idxChat], msg)
			m.syncMutex[m.idxChat].Unlock()
		default:
		}
	}

	m.textInput = ""

	return m
}

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

func initialMainBar() chatStyles {
	items := []list.Item{}
	delegate := defaultDelegate()
	list := newList(items, delegate)

	return chatStyles{
		box:   lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Width(100).Height(8),
		body:  list,
		input: lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, false, true, false).
			Width(98).
			Height(1),
		color: "white",
	}
}

func defaultDelegate() list.DefaultDelegate {
	delegate := list.NewDefaultDelegate()
	delegate.SetSpacing(0)
	delegate.ShowDescription = false
	return delegate
}

func newList(items []list.Item, delegate list.DefaultDelegate) list.Model {
	l := list.New(items, delegate, 98, 10)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	l.SetShowPagination(false)
	return l
}

func InitialModel() *model {
	items := []list.Item{}
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
		chat:      make([]chan api.Msg, 0),
		lg: loginSystem{
			isLoggin:  true,
			idxLoggin: 0,
			chat:      make([]string, 4),
		},
		sendMsg: sendMsgToLoginSystem,
		channels: make([]api.Channel, 0),
		idxChat: 0,
		idxMsgs: 0,
	}
	return &m
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

func syncLists(m model) model {
	if m.lg.isLoggin {
		menu := []list.Item{
			menuItem{id: "1", title: talkToLogin1},
			menuItem{id: "2", title: "Voce:" + m.lg.chat[0]},
			menuItem{id: "3", title: talkToLogin2},
			menuItem{id: "4", title: "Voce:" + m.lg.chat[1]},
			menuItem{id: "5", title: talkToLogin3},
			menuItem{id: "6", title: "Voce:" + m.lg.chat[2]},
			menuItem{id: "7", title: talkToLogin4},
			menuItem{id: "8", title: "Voce:" + m.lg.chat[3]},
		}

		end := min((m.lg.idxLoggin+1)*2-1, 8)
		m.mainMenu.body.SetItems(menu[:end])
	} else {
		menu := []list.Item{}
		m.syncMutex[m.idxChat].Lock()
		for idx := len(*m.msgsShow[m.idxChat]) - 1; idx >= 0; idx-- {
			menu = append(menu, menuItem{
				id:    strconv.Itoa(idx),
				title: (*m.msgsShow[m.idxChat])[idx].Text,
			})
		}
		m.mainMenu.body.SetItems(menu)
		m.syncMutex[m.idxChat].Unlock()
	}

	items := []list.Item{}

	for idx := range m.channels {
		items = append(items, menuItem{
			id:    strconv.Itoa(idx),
			title: m.channels[idx].ChannelName,
		})
	}

	m.listModel.SetItems(items)
	return m
}

func renderMain(m model) string {
	m.mainMenu.body.Title = "Chat — Geral"
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
	m = syncLists(m)
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
		case "enter":
			m.idxChat = m.listModel.Index()
			return syncLists(m), nil
		case "a":
			m.lg.isLoggin = true
			m.lg.idxLoggin = 0
			m.sendMsg = sendMsgToLoginSystem
			return swapFocus(m)
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
	if (m.lg.isLoggin && m.lg.idxLoggin == 4) {
		tryChn, isRunning := api.ConnectTo(m.lg.chat[0], m.lg.chat[1])
		if !isRunning {
			m.lg.chat = make([]string, 4)
			m.lg.idxLoggin = 0
			return m, nil
		}
		tryChn.Reader = bufio.NewReader(tryChn.Connection)
		if !api.Login(m.lg.chat[2], m.lg.chat[3], &tryChn) {
			m.lg.chat = make([]string, 4)
			m.lg.idxLoggin = 0
			return m, nil
		}
		m.channels = append(m.channels, tryChn)
		m.idxChat = len(m.channels) - 1
		m.lg.isLoggin = false
		m.lg.idxLoggin = 0
		m.msgs = append(m.msgs, make(chan api.Msg, 8))
		m.syncMutex = append(m.syncMutex, &sync.Mutex{})
		m.msgsShow = append(m.msgsShow, &[]api.Msg{})
		m.sendMsg = sendMsgToServerSystem
		m = syncLists(m)
		go api.RoutineReadMsg(&m.channels[m.idxChat], m.msgsShow[m.idxChat], m.syncMutex[m.idxChat], m.P)
		go api.RoutineSendMsg(&m.channels[m.idxChat], m.msgs[m.idxChat])
		return m, nil
	}
	return cChat(msg, m)
}

func cChat(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			return swapFocus(m)
		case "enter":
			if len(m.textInput) > 0 {
				
				m = m.sendMsg(api.Msg{
					Status: true,
					Text:   m.textInput,
				}, m)
				var cmd tea.Cmd
				m.mainMenu.body, cmd = m.mainMenu.body.Update(msg)
				m = syncLists(m)
				return m, cmd
			}
			return m, nil
		case "backspace":
			if len(m.textInput) > 0 {
				m.textInput = m.textInput[:len(m.textInput)-1]
			}
			return m, nil
		case "up", "down":
			var cmd tea.Cmd
			m.mainMenu.body, cmd = m.mainMenu.body.Update(msg)
			return m, cmd
		default:
			if msg.String() != " " || len(m.textInput) > 0 {
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
	m = syncLists(m)
	return m, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}


func max(a, b int) int {
	if a < b {
		return b
	}
	return a
}
