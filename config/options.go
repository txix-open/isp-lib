package config

type LocalOption func(l *Config)

func WithReadingFromFile(file string) LocalOption {
	return func(l *Config) {
		l.file = file
	}
}

func WithEnvPrefix(prefix string) LocalOption {
	return func(l *Config) {
		l.envPrefix = prefix
	}
}

func WithValidator(validator Validator) LocalOption {
	return func(l *Config) {
		l.validator = validator
	}
}
