package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/script-wizards/spells/internal/engine"
	"github.com/script-wizards/spells/internal/model"
)

type Model struct {
	engine    *engine.Engine
	session   *model.Session
	sessionID int64
}

func NewModel(eng *engine.Engine, sessionID int64) (Model, error) {
	session, err := model.GetSession(eng.DB, sessionID)
	if err != nil {
		return Model{}, err
	}

	m := Model{
		engine:    eng,
		session:   session,
		sessionID: sessionID,
	}

	if eng.EventBus != nil {
		eng.EventBus.Subscribe("TurnAdvanced", func(event engine.Event) {
			if turnEvent, ok := event.(engine.TurnAdvanced); ok && turnEvent.SessionID == sessionID {
				m.session.CurrentTurn = turnEvent.NewTurn
			}
		})
	}

	return m, nil
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeySpace:
			if m.engine != nil && m.sessionID > 0 {
				m.engine.Advance(m.sessionID, 1)
			}
		}
	}
	return m, nil
}

func (m Model) View() string {
	turnInfo := "Turn: Not loaded"
	if m.session != nil {
		turnInfo = fmt.Sprintf("Turn: %d", m.session.CurrentTurn)
	}

	return "Spells Tracking TUI\n\n" +
		turnInfo + "\n\n" +
		"Placeholder panes:\n" +
		"- Sessions\n" +
		"- Characters\n" +
		"- Spells\n\n" +
		"Press Space to advance turn, Ctrl+C to quit"
}
