package computation

import "github.com/strider2038/key-value-database/internal/database/querylang"

type Command struct {
	ID        querylang.CommandID
	Arguments []string
}
