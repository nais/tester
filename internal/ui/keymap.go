package ui

import "github.com/charmbracelet/bubbles/key"

type KeyMapOverview = struct {
	Down, Up, Quit key.Binding
}

var KeyMap = KeyMapOverview{
	// next: key.NewBinding(
	// 	key.WithKeys("tab"),
	// 	key.WithHelp("tab", "next"),
	// ),
	// prev: key.NewBinding(
	// 	key.WithKeys("shift+tab"),
	// 	key.WithHelp("shift+tab", "prev"),
	// ),
	// add: key.NewBinding(
	// 	key.WithKeys("ctrl+n"),
	// 	key.WithHelp("ctrl+n", "add an editor"),
	// ),
	// remove: key.NewBinding(
	// 	key.WithKeys("ctrl+w"),
	// 	key.WithHelp("ctrl+w", "remove an editor"),
	// ),
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("up", "scroll up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("down", "scroll down"),
	),
	Quit: key.NewBinding(
		key.WithKeys("esc", "ctrl+c", "q"),
		key.WithHelp("esc", "quit"),
	),
}
