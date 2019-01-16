package database

type Option func(rdc *RxDbClient)

func WithSchemaEnsuring() Option {
	return func(rdc *RxDbClient) {
		rdc.ensSchema = true
	}
}

func WithMigrationsEnsuring() Option {
	return func(rdc *RxDbClient) {
		rdc.ensMigrations = true
	}
}

func WithInitializingErrorHandler(eh errorHandler) Option {
	return func(rdc *RxDbClient) {
		rdc.eh = eh
	}
}

func WithInitializingHandler(handler initHandler) Option {
	return func(rdc *RxDbClient) {
		rdc.initHandler = handler
	}
}

func WithSchemaAutoInjecting() Option {
	return func(rdc *RxDbClient) {
		rdc.schemaInjecting = true
	}
}
