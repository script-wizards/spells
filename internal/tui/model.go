package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/script-wizards/spells/internal/engine"
	"github.com/script-wizards/spells/internal/model"
	"github.com/script-wizards/spells/internal/search"
)

type Mode int

const (
	NormalMode Mode = iota
	SearchMode
)

type Model struct {
	engine        *engine.Engine
	session       *model.Session
	sessionID     int64
	mode          Mode
	searchQuery   string
	searchIndex   search.Index
	searchResults []model.NPC
}

func NewModel(eng *engine.Engine, sessionID int64) (Model, error) {
	session, err := model.GetSession(eng.DB, sessionID)
	if err != nil {
		return Model{}, err
	}

	// Initialize search index with NPC names
	npcNames, err := model.GetAllNPCNames(eng.DB)
	if err != nil {
		return Model{}, err
	}
	searchIndex := search.BuildIndex(npcNames)

	m := Model{
		engine:      eng,
		session:     session,
		sessionID:   sessionID,
		mode:        NormalMode,
		searchQuery: "",
		searchIndex: searchIndex,
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
		switch m.mode {
		case NormalMode:
			switch msg.Type {
			case tea.KeyCtrlC:
				return m, tea.Quit
			case tea.KeySpace:
				if m.engine != nil && m.sessionID > 0 {
					m.engine.Advance(m.sessionID, 1)
				}
			default:
				if msg.Type == tea.KeyRunes {
					switch string(msg.Runes) {
					case "/":
						m.mode = SearchMode
						m.searchQuery = ""
						m.searchResults = nil
					}
				}
			}
		case SearchMode:
			switch msg.Type {
			case tea.KeyCtrlC:
				return m, tea.Quit
			case tea.KeyEsc:
				m.mode = NormalMode
				m.searchQuery = ""
				m.searchResults = nil
			case tea.KeyEnter:
				m.mode = NormalMode
			case tea.KeyBackspace:
				if len(m.searchQuery) > 0 {
					m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
					m.updateSearchResults()
				}
			default:
				if msg.Type == tea.KeyRunes {
					m.searchQuery += string(msg.Runes)
					m.updateSearchResults()
				}
			}
		}
	}
	return m, nil
}

func (m *Model) updateSearchResults() {
	if m.searchQuery == "" {
		m.searchResults = nil
		return
	}

	if m.engine != nil && m.engine.DB != nil {
		results, err := model.SearchNPC(m.engine.DB, m.searchIndex, m.searchQuery, 5)
		if err == nil {
			m.searchResults = results
		}
	}
}

func (m Model) View() string {
	turnInfo := "Turn: Not loaded"
	if m.session != nil {
		turnInfo = fmt.Sprintf("Turn: %d", m.session.CurrentTurn)
	}

	var view strings.Builder
	view.WriteString("Spells Tracking TUI\n\n")
	view.WriteString(turnInfo + "\n\n")

	if m.mode == SearchMode {
		view.WriteString("Search Mode (ESC to exit, Enter to select)\n")
		view.WriteString(fmt.Sprintf("Query: %s\n\n", m.searchQuery))

		if len(m.searchResults) > 0 {
			view.WriteString("NPCs found:\n")
			for i, npc := range m.searchResults {
				if i >= 5 { // Only show top 5
					break
				}
				location := ""
				if npc.Location != nil {
					location = fmt.Sprintf(" (%s)", *npc.Location)
				}
				view.WriteString(fmt.Sprintf("  %d. %s%s\n", i+1, npc.Name, location))
			}
		} else if m.searchQuery != "" {
			view.WriteString("No NPCs found.\n")
		}
	} else {
		view.WriteString("Placeholder panes:\n")
		view.WriteString("- Sessions\n")
		view.WriteString("- Characters\n")
		view.WriteString("- Spells\n\n")
		view.WriteString("Press '/' for NPC search, Space to advance turn, Ctrl+C to quit")
	}

	return view.String()
}
