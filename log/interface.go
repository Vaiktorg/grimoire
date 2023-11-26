package log

type ILogger interface {
	ServiceName() string
	NewServiceLogger(config Config) IServiceLogger
	Services() map[string]ILogger

	TRACE(info string, obj ...any)
	INFO(info string, obj ...any)
	DEBUG(procStep string, obj ...any)
	WARN(warn string, obj ...any)
	ERROR(errMsg error, obj ...any)
	FATAL(breakage string)

	Println(in ...any)
	Printf(str string, data ...any)

	Messages(Pagination) []Log
	BatchLogs(logs ...Log)
	Output(func(log Log))
	TotalSent() uint64
	Close()

	HasLevel(flag Level) bool
	AddLevel(flag Level)
	ClearLevel(flag Level)
	ToggleLevel(flag Level)
}

type IServiceLogger interface {
	ServiceName() string
	NewServiceLogger(config Config) IServiceLogger

	TRACE(info string, obj ...any)
	INFO(info string, obj ...any)
	DEBUG(procStep string, obj ...any)
	WARN(warn string, obj ...any)
	ERROR(errMsg error, obj ...any)
	FATAL(breakage string)

	Println(in ...any)
	Printf(str string, data ...any)

	Messages(Pagination) []Log
	BatchLogs(logs ...Log)
	Output(func(log Log))
	TotalSent() uint64

	HasLevel(flag Level) bool
	AddLevel(flag Level)
	ClearLevel(flag Level)
	ToggleLevel(flag Level)
}
