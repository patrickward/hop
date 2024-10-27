package processors

// HTMLProcessor defines the interface for processing HTML content
type HTMLProcessor interface {
	Process(html string) (string, error)
}

// CompositeProcessor combines multiple HTML processors into one. During the processing stage,
// it will apply each processor in the order they were added.
type CompositeProcessor struct {
	processors []HTMLProcessor
}

// NewCompositeProcessor creates a new CompositeProcessor with the given processors.
func NewCompositeProcessor(processors ...HTMLProcessor) *CompositeProcessor {
	return &CompositeProcessor{processors: processors}
}

// Process applies all processors in order to the given HTML string.
func (c *CompositeProcessor) Process(html string) (string, error) {
	var err error
	for _, processor := range c.processors {
		html, err = processor.Process(html)
		if err != nil {
			return "", err
		}
	}
	return html, nil
}
