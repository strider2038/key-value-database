package parsing

import (
	"fmt"
	"strings"
	"unicode"
)

type event int

const (
	foundSymbolEvent event = iota
	foundSpaceEvent

	eventsCount
)

type state int

const (
	initialState state = iota
	wordState
	spaceState

	statesCount
)

type transition struct {
	jump   func(rune) state
	action func()
}

type stateMachine struct {
	transitions [statesCount][eventsCount]transition
	state       state

	tokens []string
	s      strings.Builder
}

func newStateMachine() *stateMachine {
	sm := &stateMachine{state: initialState, tokens: make([]string, 0)}

	sm.transitions = [statesCount][eventsCount]transition{
		initialState: {
			foundSymbolEvent: {jump: sm.addSymbol},
			foundSpaceEvent:  {jump: sm.skipSpace},
		},
		wordState: {
			foundSymbolEvent: {jump: sm.addSymbol},
			foundSpaceEvent:  {jump: sm.skipSpace, action: sm.addToken},
		},
		spaceState: {
			foundSymbolEvent: {jump: sm.addSymbol},
			foundSpaceEvent:  {jump: sm.skipSpace},
		},
	}

	return sm
}

func (m *stateMachine) Parse(raw string) ([]string, error) {
	for _, c := range raw {
		switch {
		case unicode.IsSpace(c):
			m.processEvent(foundSpaceEvent, c)
		case isSymbol(c):
			m.processEvent(foundSymbolEvent, c)
		default:
			return nil, fmt.Errorf("%w: %q", ErrUnexpectedSymbol, c)
		}
	}

	m.processEvent(foundSpaceEvent, ' ')

	return m.tokens, nil
}

func (m *stateMachine) processEvent(event event, symbol rune) {
	transition := m.transitions[m.state][event]
	m.state = transition.jump(symbol)
	if transition.action != nil {
		transition.action()
	}
}

func (m *stateMachine) addSymbol(r rune) state {
	m.s.WriteRune(r)

	return wordState
}

func (m *stateMachine) skipSpace(r rune) state {
	return spaceState
}

func (m *stateMachine) addToken() {
	m.tokens = append(m.tokens, m.s.String())
	m.s.Reset()
}

func isSymbol(c rune) bool {
	return c >= '0' && c <= '9' ||
		c >= 'a' && c <= 'z' ||
		c >= 'A' && c <= 'Z' ||
		c == '*' || c == '_' || c == '/'
}
