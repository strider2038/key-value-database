package parsing

type Parser struct{}

func NewParser() *Parser {
	return &Parser{}
}

func (p *Parser) ParseCommand(command string) ([]string, error) {
	sm := newStateMachine()

	return sm.Parse(command)
}
