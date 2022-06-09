package hub

// Cloner - structure that clones itself.
type Cloner interface {
	Clone() Cloner
}
