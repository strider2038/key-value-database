package querylang

type CommandID int

func (c CommandID) String() string {
	switch c {
	case CommandGet:
		return "GET"
	case CommandSet:
		return "SET"
	case CommandDel:
		return "DET"
	default:
		return ""
	}
}

const (
	CommandUnknown CommandID = iota
	CommandSet
	CommandGet
	CommandDel
)

type Command struct {
	id        CommandID
	arguments []string
}

func (c *Command) ID() CommandID       { return c.id }
func (c *Command) Arguments() []string { return c.arguments }

func NewCommand(id CommandID, arguments ...string) *Command {
	return &Command{id: id, arguments: arguments}
}
