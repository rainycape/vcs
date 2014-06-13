package vcs

// Interface is the interface implemented by VCS
// systems. To register a new one, use Register.
type Interface interface {
	Cmd() string
	Dir() string
	Head() string
	Revision(id string) []string
	History(since string) []string
	Checkout(rev string) []string
	Clone(src string, dst string) []string
	Update() []string
	ParseRevisions(since string, data []byte) ([]*Revision, error)
	Branches() []string
	Tags() []string
	ParseBranches(data []byte) ([]*Branch, error)
	ParseTags(data []byte) ([]*Tag, error)
}

var (
	interfaces []Interface
)

// Register adds a new Interface to the list of supported VCS interfaces.
// To allow overriding the defaults, Register adds the new Interface at
// the start of the internal Interface list.
func Register(iface Interface) {
	// Put at the front, to allow overriding defaults
	interfaces = append([]Interface{iface}, interfaces...)
}
