package main

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ItemList struct {
	items []string
}

type model struct {
	active   int
	items    []string
	viewport viewport.Model
	ready    bool
	textarea textarea.Model
}

func (m model) buildContentView() string {
	m.viewport.Height = m.viewport.Height - 10
	m.viewport.Width = 40

	return m.viewport.View()
}

func (m model) buildMenuView() string {
	menu := lipgloss.NewStyle().
		PaddingRight(5).
		PaddingLeft(5).
		Align(lipgloss.Left)

	collections := ""

	for v, i := range m.items {
		if m.active == v {
			if collections == "" {
				s := lipgloss.NewStyle().Foreground(lipgloss.Color("77"))
				collections = lipgloss.JoinVertical(lipgloss.Left, s.Render(i))
			} else {
				s := lipgloss.NewStyle().Foreground(lipgloss.Color("77"))
				collections = lipgloss.JoinVertical(lipgloss.Left, collections, s.Render(i))
			}
		} else {
			if collections == "" {
				s := lipgloss.NewStyle()
				collections = lipgloss.JoinVertical(lipgloss.Left, s.Render(i))
			} else {
				s := lipgloss.NewStyle()
				collections = lipgloss.JoinVertical(lipgloss.Left, collections, s.Render(i))
			}
		}
	}
	collections = menu.Render(collections)

	return collections
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if !m.ready {
			m.viewport = viewport.New(20, 10)
			m.viewport.HighPerformanceRendering = false
			m.viewport = viewport.New(msg.Width, msg.Height)
			m.ready = true
		} else {
			m.viewport.Width = msg.Width - 10
			m.viewport.Height = msg.Height - 10
		}
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case tea.KeyCtrlC.String():
			return m, tea.Quit
		case tea.KeyCtrlN.String():
			m.active++
			if m.active >= len(m.items) {
				m.active = len(m.items) - 1
			}
		case tea.KeyCtrlP.String():
			m.active--
			if m.active <= 0 {
				m.active = 0
			}
		case tea.KeyCtrlR.String():
			url := m.items[m.active]
			resp, err := http.Get(url)
			if err != nil {
				return m, tea.Quit
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				fmt.Println("Error reading response:", err)
				return m, tea.Quit
			}

			m.viewport.SetContent(string(body))
		case tea.KeyEnter.String():
			if !m.textarea.Focused() {
				m.textarea.Focus()
				return m, cmd
			}
		case tea.KeyEscape.String():
			if m.textarea.Focused() {
				m.textarea.Blur()
			}
		}
	}
	if !m.textarea.Focused() {
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	m.textarea, cmd = m.textarea.Update(msg)
	cmds = append(cmds, cmd)

	return m, cmd
}

func (m model) View() string {
	outerStyle := lipgloss.NewStyle().
		BorderTop(true).
		BorderBottom(true).
		BorderRight(true).
		BorderLeft(true).
		BorderStyle(lipgloss.RoundedBorder())

	cvh := lipgloss.Height(m.buildContentView())
	m.textarea.SetHeight(cvh)
	return outerStyle.Render(
		lipgloss.JoinHorizontal(
			lipgloss.Left,
			m.buildMenuView(),
			m.textarea.View(),
			lipgloss.NewStyle().
				BorderLeft(true).
				BorderStyle(lipgloss.NormalBorder()).
				Render(m.buildContentView()),
		),
	)
}

func main() {
	ti := textarea.New()
	ti.Placeholder = "Request Body"
	ti.Focus()
	ti.Blur()

	m := model{
		active: 0,
		items: []string{
			"http://example.com",
			"https://jsonplaceholder.typicode.com/todos/1",
			"https://jsonplaceholder.typicode.com/posts/1",
		},
		textarea: ti,
	}

	if _, err := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion()).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
