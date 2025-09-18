package logging

import "io"

type Config struct {
	Level       string
	JSON        bool
	AddSource   bool
	Writer      io.Writer
	Service     string
	Version     string
	Environment string
	SetDefault  bool
	TimeFormat  string
}
