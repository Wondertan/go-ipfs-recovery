package recovery

// Strategy defines verbosity level for Node recovery.
// Higher verbosity potentially uses more storage space.
type Strategy string

const (

	// Recover all the Nodes possible. Redundant and Data Nodes altogether.
	// Most storage usage.
	All Strategy = "all"

	// Recover only Data Nodes. Only user data, ignores Redundant Nodes.
	Data Strategy = "data"

	// Recover only requested Nodes. Actual Nodes user asked for, might be just small portion of all the Data.
	Requested Strategy = "requested"
)

func (s Strategy) All() bool {
	return s == All
}

func (s Strategy) Data() bool {
	return s == All || s == Data
}

func (s Strategy) Requested() bool {
	return s == All || s == Data || s == Requested
}
