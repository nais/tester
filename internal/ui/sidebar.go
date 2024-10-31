package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const sidebarWidth = 30

var unselectedStyle = lipgloss.NewStyle()

var selectedStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#FFF")).
	Background(lipgloss.Color("#333"))

type Sidebar struct {
	IsOpen   bool
	data     string
	viewport viewport.Model
	ctx      *ProgramContext

	files []string
	index int
}

func NewSidebar(ctx *ProgramContext) Sidebar {
	return Sidebar{
		IsOpen: true,
		viewport: viewport.Model{
			Width:  0,
			Height: 0,
		},
		ctx: ctx,
		files: []string{
			"team/confirm_team_deletion.lua",
			"team/create_team.lua",
			"ingress_types.lua",
			"list_users.lua",
			"manage_team_members.lua",
			"request_team_deletion.lua",
			"secrets.lua",
			"status_for_applications.lua",
			"status_for_jobs.lua",
			"team_applications.lua",
			"team_viewer.lua",
			"unauthenticated_list_users.lua",
			"update_team_environment.lua",
			"update_team.lua",
		},
	}
}

func (s Sidebar) Update(msg tea.Msg) (Sidebar, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, KeyMap.Down):
			s.index++
			// s.viewport.HalfViewDown()

		case key.Matches(msg, KeyMap.Up):
			s.index--
			// s.viewport.HalfViewUp()
		}
	}

	if s.index > len(s.files)-1 {
		s.index = len(s.files) - 1
	} else if s.index < 0 {
		s.index = 0
	}

	sb := &strings.Builder{}
	for i, f := range s.files {

		name := f
		if len(name) > s.GetSidebarContentWidth() {
			name = name[:s.GetSidebarContentWidth()-3] + "..."
		}
		if i == s.index {
			sb.WriteString(selectedStyle.Render(name))
		} else {
			sb.WriteString(unselectedStyle.Render(name))
		}
		sb.WriteRune('\n')
	}

	s.data = sb.String()
	s.viewport.SetContent(s.data)

	return s, nil
}

func (s Sidebar) Init() tea.Cmd {
	return nil
}

func (s Sidebar) View() string {
	if !s.IsOpen {
		return ""
	}

	height := s.ctx.MainContentHeight
	style := lipgloss.NewStyle().
		Padding(0, 2, 0, 0).
		BorderRight(true).
		BorderStyle(lipgloss.Border{
			Top:         "",
			Bottom:      "",
			Left:        "",
			Right:       "│",
			TopLeft:     "",
			TopRight:    "",
			BottomRight: "",
			BottomLeft:  "",
		}).
		BorderForeground(lipgloss.AdaptiveColor{Light: "013", Dark: "008"}).
		Height(height).
		MaxHeight(height).
		Width(sidebarWidth).
		MaxWidth(sidebarWidth)

	if s.data == "" {
		return style.Align(lipgloss.Center).Render(
			lipgloss.PlaceVertical(height, lipgloss.Center, "No tests"),
		)
	}

	return style.Render(lipgloss.JoinVertical(
		lipgloss.Top,
		s.viewport.View(),
		lipgloss.NewStyle().
			Height(2).
			Bold(true).
			Foreground(lipgloss.Color("#333")).
			Render(fmt.Sprintf("%d%%", int(s.viewport.ScrollPercent()*100))),
	))
}

func (s *Sidebar) SetContent(filenames []string) {
	s.files = filenames
	if s.index > len(s.files)-1 {
		s.index = len(s.files) - 1
	}

	if s.index < 0 {
		s.index = 0
	}
}

func (s *Sidebar) UpdateProgramContext(ctx *ProgramContext) {
	if ctx == nil {
		return
	}
	s.ctx = ctx
	s.viewport.Height = s.ctx.MainContentHeight - 1
	s.viewport.Width = s.GetSidebarContentWidth()
}

func (s *Sidebar) GetSidebarContentWidth() int {
	return sidebarWidth - 2*2 - 1
}
