package pulse

type Batcher interface {
	// Flush the batcher
	Flush() error
}
