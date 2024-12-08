package conf

import (
	"flag"
)

func NewFlagSet(appName string, errorHandling flag.ErrorHandling) *flag.FlagSet {
	fs := flag.NewFlagSet("myapp", errorHandling)
	BasicFlags(fs)
	return fs
}

// BasicFlags sets up the common flags most apps will need
func BasicFlags(fs *flag.FlagSet) {
	fs.String("config", "config.json", "Path to config file")
	fs.Bool("version", false, "Show version and exit")
}

//// ApplyFlagOverrides ensures that if a basic flag is set, it overrides the config
//func ApplyFlagOverrides(cfg *HopConfig, fs *flag.FlagSet) {
//	//if fs.Lookup("debug").Value.String() == "true" {
//	//	cfg.App.Debug = true
//	//}
//	//if fs.Lookup("env").Value.String() != "" {
//	//	cfg.App.Environment = fs.Lookup("env").Value.String()
//	//}
//}
