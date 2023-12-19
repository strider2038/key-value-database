package querylang

type CommandID int

const (
	CommandUnknown CommandID = iota
	CommandSet
	CommandGet
	CommandDel

	commandLast
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
