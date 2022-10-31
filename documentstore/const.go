package documentstore

const FileIDLength = 10
const DirIDLength = 10

type FilePermission int
type DirPermission int

const (
	PublicFile FilePermission = iota
	PrivateFile
	SecretFile

	PublicDir DirPermission = iota
	PrivateDir
	SecretDir
)
