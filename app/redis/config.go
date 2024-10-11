package redis

import "fmt"

type Config struct {
	Dir        string
	Dbfilename string
}

var cfg *Config //singleton

func SetConfig(args []string) error {
	if len(args) <= 1 {
		return nil
	}
	if len(args)%2 != 0 {
		return fmt.Errorf("invalid number of arguments. Expected key-value pairs")
	}

	if cfg == nil {
		cfg = &Config{}
	}

	for i := 0; i < len(args); i += 2 {
		key := args[i]
		value := args[i+1]
		switch key {
		case "--dir":
			cfg.Dir = value
		case "--dbfilename":
			cfg.Dbfilename = value
		default:
			return fmt.Errorf("unknown config")
		}
	}

	return nil
}

func GetConfig() Config {
	return *cfg // would this return a new copy?
}
