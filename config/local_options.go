package config

type LocalOption func(l *Local)

func WithReadingFromFile(file string) LocalOption {
	return func(l *Local) {
		l.file = file
	}
}

func WithEnvPrefix(prefix string) LocalOption {
	return func(l *Local) {
		l.envPrefix = prefix
	}
}

func WithValidator(validator Validator) LocalOption {
	return func(l *Local) {
		l.validator = validator
	}
}
