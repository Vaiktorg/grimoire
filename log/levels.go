package log

// Level
/*
   Trace - Only when I would be "tracing" the code and trying to find one part of a function specifically.
   Debug - Information that is diagnostically helpful to people more than just developers (IT, sysadmins, etc.).
   Info - Generally useful information to src (services start/stop, configuration assumptions, etc). Info I want to always have available but usually don't care about under normal circumstances. This is my out-of-the-box config level.
   Warn - Anything that can potentially cause application oddities, but for which I am automatically recovering. (Such as switching from a primary to backup server, retrying an operation, missing secondary data, etc.)
   Error - Any error which is fatal to the operation, but not the services or application (can't open a required file, missing data, etc.). These errors will force user (administrator, or direct user) intervention. These are usually reserved (in my apps) for incorrect connection strings, missing serv, etc.
   Fatal - Any error that is forcing a shutdown of the services or application to prevent data loss (or further data loss). I reserve these only for the most heinous errors and situations where there is guaranteed to have been data corruption or loss.
*/

type Level uint8

const (
	LevelNull  Level = iota // ∅	- disable line
	LevelTrace              // trace - white
	LevelDebug              // debug	- grey
	LevelInfo               // info	- blue
	LevelWarn               // warn	- orange
	LevelError              // error	- red
	LevelFatal              // fatal	- black
)

func LevelFromString(str string) Level {
	lvl := LevelNull
	return lvl.Level(str)
}

func (l *Level) Set(flag Level)      { *l = *l | flag }
func (l *Level) Clear(flag Level)    { *l = *l &^ flag }
func (l *Level) Toggle(flag Level)   { *l = *l ^ flag }
func (l *Level) Has(flag Level) bool { return *l&flag != 0 }
func (l *Level) Is(flag Level) bool  { return *l == flag }

func (l *Level) String() string {
	switch *l {
	case LevelTrace:
		return "TRACE"
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	case LevelFatal:
		return "FATAL"
	default: // LevelNull
		return "∅"
	}
}

func (*Level) Level(str string) Level {
	switch str {
	case "TRACE":
		return LevelTrace
	case "DEBUG":
		return LevelDebug
	case "INFO ":
		return LevelInfo
	case "WARN ":
		return LevelWarn
	case "ERROR":
		return LevelError
	case "FATAL":
		return LevelFatal
	default:
		return LevelNull
	}
}

// ==================================================
