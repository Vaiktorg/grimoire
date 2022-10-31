package log

var lggr = new(Logger)

func Init() {
	lggr.inChan = make(chan Log)

	// Default levels
	lggr.LogLevels = LevelTrace | LevelInfo | LevelDebug | LevelError

	go lggr.handlePersistence()
}

func TRACE(info string, obj ...interface{}) {
	lggr.newMsg(info, obj, LevelTrace)
}
func INFO(info string, obj ...interface{}) {
	lggr.newMsg(info, obj, LevelInfo)
}
func DEBUG(procstep string, obj ...interface{}) {
	lggr.newMsg(procstep, obj, LevelDebug)
}
func WARN(warn string, obj ...interface{}) {
	lggr.newMsg(warn, obj, LevelWarn)
}
func ERROR(errmsg error, obj ...interface{}) {
	lggr.newMsg(errmsg.Error(), obj, LevelError)
}
func FATAL(breakage string, obj ...interface{}) {
	lggr.newMsg(breakage, obj, LevelFatal)
}
func Close() {
	lggr.Close()
}
