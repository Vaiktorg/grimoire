package log

type ISimLogger interface {
	TRACE(info string, obj ...any)
	INFO(info string, obj ...any)
	DEBUG(procStep string, obj ...any)
	WARN(warn string, obj ...any)
	ERROR(errMsg string, obj ...any) string
	FATAL(breakage string)
}

type ILogger interface {
	ISimLogger
	ServiceName() string

	Messages(Pagination) []Log

	NewServiceLogger(config *Config) ILogger
	Services() map[string]ILogger

	Println(in ...any)
	Printf(str string, data ...any)

	TotalSent() uint64
	BatchLogs(...Log)
	Output(func(log Log) error)
	Close()
}
