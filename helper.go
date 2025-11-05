package logging

import "github.com/rs/zerolog"

func getLevel(level string) (zerolog.Level, error) {
	l, err := zerolog.ParseLevel(level)
	if err != nil {
		return zerolog.DebugLevel, err
	}
	return l, nil
}
