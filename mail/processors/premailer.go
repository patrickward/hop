package processors

import "github.com/vanng822/go-premailer/premailer"

// PremailerProcessor implements HTMLProcessor using premailer
type PremailerProcessor struct {
	options *premailer.Options
}

// NewPremailerProcessor creates a new PremailerProcessor with the given options
func NewPremailerProcessor(options *premailer.Options) *PremailerProcessor {
	if options == nil {
		options = premailer.NewOptions()
	}
	return &PremailerProcessor{options: options}
}

// Process applies the premailer transformation to the given HTML string
func (p *PremailerProcessor) Process(html string) (string, error) {
	prem, err := premailer.NewPremailerFromString(html, p.options)
	if err != nil {
		return "", err
	}
	return prem.Transform()
}
