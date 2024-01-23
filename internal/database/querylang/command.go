package querylang

type CommandID int

func (c CommandID) String() string {
	switch c {
	case CommandGet:
		return "GET"
	case CommandSet:
		return "SET"
	case CommandDel:
		return "DEL"
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
	seqID     uint64
	id        CommandID
	arguments []string
}

func (c *Command) SeqID() uint64       { return c.seqID }
func (c *Command) ID() CommandID       { return c.id }
func (c *Command) Arguments() []string { return c.arguments }

func (c *Command) IsReadOperation() bool {
	return c.id == CommandGet
}

func NewCommand(seqID uint64, id CommandID, arguments ...string) *Command {
	return &Command{seqID: seqID, id: id, arguments: arguments}
}
