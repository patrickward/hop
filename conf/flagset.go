package conf

import (
	"flag"
)

func NewFlagSet(appName string, errorHandling flag.ErrorHandling) *flag.FlagSet {
	fs := flag.NewFlagSet("myapp", errorHandling)
	BasicFlags(fs)
	return fs
}

//func ParseAndApplyFlags(cfg *BaseConfig, fs *flag.FlagSet) error {
//	if err := fs.Parse(os.Args[1:]); err != nil {
//		return err
//	}
//	ApplyFlagOverrides(cfg, fs)
//	return nil
//}

//// SetupFlags registers basic flags and allows for custom flag registration
//type SetupFlags func(*flag.FlagSet)

// BasicFlags sets up the common flags most apps will need
func BasicFlags(fs *flag.FlagSet) {
	fs.String("config", "config.json", "Path to config file")
	fs.String("env", "development", "Environment (development, staging, production)")
	fs.Bool("debug", false, "Enable debug mode")
	fs.Bool("version", false, "Show version and exit")
}

// ApplyFlagOverrides ensures that if a basic flag is set, it overrides the config
func ApplyFlagOverrides(cfg *BaseConfig, fs *flag.FlagSet) {
	if fs.Lookup("debug").Value.String() == "true" {
		cfg.App.Debug = true
	}
	if fs.Lookup("env").Value.String() != "" {
		cfg.App.Environment = fs.Lookup("env").Value.String()
	}
}
