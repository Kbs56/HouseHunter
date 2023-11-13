package main

// A simple example demonstrating the use of multiple text input components
// from the Bubbles component library.

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	focusedStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("#93E9BE"))
	blurredStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	cursorStyle         = focusedStyle.Copy()
	noStyle             = lipgloss.NewStyle()
	helpStyle           = blurredStyle.Copy()
	cursorModeHelpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	focusedButton       = focusedStyle.Copy().Render("[ Submit ]")
	blurredButton       = fmt.Sprintf("[ %s ]", blurredStyle.Render("Submit"))
)

type model struct {
	focusIndex    int
	inputs        []textinput.Model
	cursorMode    cursor.Mode
	spinner       spinner.Model
	typing        bool
	loading       bool
	SearchResults string
}

type GotHouses struct {
	Data string
}

func initialModel() model {
	m := model{
		inputs: make([]textinput.Model, 8),
	}

	var t textinput.Model
	for i := range m.inputs {
		t = textinput.New()
		t.Cursor.Style = cursorStyle
		t.CharLimit = 32

		switch i {
		case 0:
			t.Placeholder = "Define your search area(s) Ex.(Dallas,Houston) or (75204,Austin)"
			t.Focus()
			t.PromptStyle = focusedStyle
			t.TextStyle = focusedStyle
		case 1:
			t.Placeholder = "Minimum Price"
			t.PromptStyle = focusedStyle
			t.TextStyle = focusedStyle
		case 2:
			t.Placeholder = "Maximum Price"
			t.PromptStyle = focusedStyle
			t.TextStyle = focusedStyle
		case 3:
			t.Placeholder = "Bed Minimum"
			t.PromptStyle = focusedStyle
			t.TextStyle = focusedStyle
		case 4:
			t.Placeholder = "Bath Minimum"
			t.PromptStyle = focusedStyle
			t.TextStyle = focusedStyle
		case 5:
			t.Placeholder = "Square Footage Minimum"
			t.PromptStyle = focusedStyle
			t.TextStyle = focusedStyle
		case 6:
			t.Placeholder = "Looking to Buy or Rent?"
			t.PromptStyle = focusedStyle
			t.TextStyle = focusedStyle
		case 7:
			t.Placeholder = "Specify Number of Results to see per search area"
			t.PromptStyle = focusedStyle
			t.TextStyle = focusedStyle
		}
		m.inputs[i] = t
	}
	s := spinner.New()
	s.Spinner = spinner.Dot
	m.spinner = s
	m.typing = true
	m.loading = false
	return m
}

func getData(inputs []textinput.Model) string {
	time.Sleep(time.Second * 2)
	s := strings.Builder{}
	for _, v := range inputs {
		s.WriteString(v.Value() + "\n")
	}
	return s.String()
}

func (m model) fetchResults(inputs []textinput.Model) tea.Cmd {
	return func() tea.Msg {
		message := houseHunt(inputs)
		return GotHouses{Data: message}
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "r":
			m.typing = true
			m.loading = false
			m := initialModel()
			return m, nil
		case "ctrl+c", "esc":
			return m, tea.Quit

		// Change cursor mode
		case "ctrl+r":
			m.cursorMode++
			if m.cursorMode > cursor.CursorHide {
				m.cursorMode = cursor.CursorBlink
			}
			cmds := make([]tea.Cmd, len(m.inputs))
			for i := range m.inputs {
				cmds[i] = m.inputs[i].Cursor.SetMode(m.cursorMode)
			}
			return m, tea.Batch(cmds...)

		// Set focus to next input
		case "tab", "shift+tab", "enter", "up", "down":
			s := msg.String()

			// Did the user press enter while the submit button was focused?
			// If so, call the api, get the results and then display the results
			if s == "enter" && m.focusIndex == len(m.inputs) {
				m.typing = false
				m.loading = true
				return m, tea.Batch(m.spinner.Tick, m.fetchResults(m.inputs))
			}

			// Cycle indexes
			if s == "up" || s == "shift+tab" {
				m.focusIndex--
			} else {
				m.focusIndex++
			}

			if m.focusIndex > len(m.inputs) {
				m.focusIndex = 0
			} else if m.focusIndex < 0 {
				m.focusIndex = len(m.inputs)
			}

			cmds := make([]tea.Cmd, len(m.inputs))
			for i := 0; i <= len(m.inputs)-1; i++ {
				if i == m.focusIndex {
					// Set focused state
					cmds[i] = m.inputs[i].Focus()
					m.inputs[i].PromptStyle = focusedStyle
					m.inputs[i].TextStyle = focusedStyle
					continue
				}
				// Remove focused state
				m.inputs[i].Blur()
				m.inputs[i].PromptStyle = noStyle
				m.inputs[i].TextStyle = noStyle
			}
			return m, tea.Batch(cmds...)
		}
	case GotHouses:
		m.SearchResults = msg.Data
		m.loading = false
		return m, nil
	}

	if m.loading {
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	// Handle character input and blinking
	cmd := m.updateInputs(msg)

	return m, cmd
}

func (m *model) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))

	// Only text inputs with Focus() set will respond, so it's safe to simply
	// update all of them here without any further logic.
	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

func (m model) View() string {
	if m.typing {

		var b strings.Builder

		for i := range m.inputs {
			b.WriteString(m.inputs[i].View())
			if i < len(m.inputs)-1 {
				b.WriteRune('\n')
			}
		}

		button := &blurredButton
		if m.focusIndex == len(m.inputs) {
			button = &focusedButton
		}
		fmt.Fprintf(&b, "\n\n%s\n\n", *button)

		b.WriteString(helpStyle.Render("cursor mode is "))
		b.WriteString(cursorModeHelpStyle.Render(m.cursorMode.String()))
		b.WriteString(helpStyle.Render(" (ctrl+r to change style)"))

		return b.String()
	}
	if m.loading {
		return fmt.Sprintf("%s Fetching your message...", m.spinner.View())
	}
	return fmt.Sprintf(
		"Here are the houses that we fetched for you: \n\n%s\nPress (ctrl+c to quit, or r to search again)",
		m.SearchResults,
	)
}

func main() {
	if _, err := tea.NewProgram(initialModel(), tea.WithAltScreen()).Run(); err != nil {
		fmt.Printf("could not start program: %s\n", err)
		os.Exit(1)
	}
}
