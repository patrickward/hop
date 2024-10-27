package processors

//import "github.com/premailer/premailer"
//
//// PremailerProcessor implements HTMLProcessor using premailer
//type PremailerProcessor struct {
//	options *premailer.Options
//}
//
//func NewPremailerProcessor(options *premailer.Options) *PremailerProcessor {
//	if options == nil {
//		options = premailer.NewOptions()
//	}
//	return &PremailerProcessor{options: options}
//}
//
//func (p *PremailerProcessor) Process(html string) (string, error) {
//	prem, err := premailer.NewPremailerFromString(html, p.options)
//	if err != nil {
//		return "", err
//	}
//	return prem.Transform()
//}
