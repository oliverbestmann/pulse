package wx

type Batcher interface {
	// Flush the batcher
	Flush()
}
