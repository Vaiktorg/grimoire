package documentstore

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"time"

	"github.com/google/uuid"

	"github.com/vaiktorg/grimoire/helpers"
)

type Metadata struct {
	Name        string `json:"name,omitempty"`
	IsProtected bool   `json:"protected,omitempty" xml:"protected,omitempty" yaml:"protected,omitempty"`
	Path        string `json:"path,omitempty"`
	Timestamp   string `json:"timestamp,omitempty"`
	Size        string `json:"size,omitempty"`
	size        int64
}

type Dir struct {
	ID    string   `json:"id" xml:"id" yaml:"id"`
	Meta  Metadata `json:"metadata" xml:"metadata" yaml:"metadata"`
	Dirs  map[string]*Dir
	Files map[string]*File

	parent     *Dir
	hasFiles   bool
	totalItems int
}

// File Represents a file
type File struct {
	ID   string   `json:"id" xml:"id" yaml:"id"`
	Meta Metadata `json:"metadata" xml:"metadata" yaml:"metadata"`
	Data []byte   `json:"data,omitempty" xml:"data,omitempty" yaml:"data,omitempty"`
}

//NewDir creates a new dir
func NewDir(path string) *Dir {
	d := &Dir{
		ID: uuid.New().String(),
		Meta: Metadata{
			Name:      path,
			Path:      path,
			Timestamp: time.Now().Format("20060102150405"),
		},
		Dirs:  make(map[string]*Dir),
		Files: make(map[string]*File),
	}

	return d
}

func (d *Dir) GenerateDirFromPath(dirPath string) error {
	tree, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return err
	}

	d.Meta.Name = filepath.Base(dirPath)
	d.Meta.Path = dirPath

	for _, elem := range tree {
		d.Meta.size += elem.Size()
		if elem.IsDir() {

			d2 := &Dir{
				// Directories
				parent: d,
				ID:     uuid.New().String(),
				Meta: Metadata{
					Name:        elem.Name(),
					Path:        filepath.Join(dirPath, elem.Name()),
					Timestamp:   helpers.MakeTimestampNum(),
					IsProtected: false,
				},
			}

			d.Dirs[d2.ID] = d2

			err = d2.GenerateDirFromPath(filepath.Join(dirPath, elem.Name()))
			if err != nil {
				return err
			}

			continue
		}
		if !elem.IsDir() {
			f := &File{
				// Files
				ID: uuid.New().String(),
				Meta: Metadata{
					Name:        elem.Name(),
					Path:        filepath.Join(dirPath, elem.Name()),
					Size:        ByteCountBinary(elem.Size()),
					Timestamp:   elem.ModTime().Format("2006-01-02_15:04:05"),
					IsProtected: false,
				},
			}
			d.Files[f.ID] = f
			d.hasFiles = true
		}
		d.totalItems++
	}

	d.Meta.Size = ByteCountBinary(d.Meta.size)

	return nil
}

// AddDir Create a directory inside structure
func (d *Dir) AddDir(name string) *Dir {
	d2 := &Dir{
		parent: d,
		ID:     uuid.New().String(),
		Meta: Metadata{
			Name:      name,
			Path:      filepath.Join(d.Meta.Path, name),
			Timestamp: helpers.MakeTimestampNum(),
		},
		Dirs:  make(map[string]*Dir),
		Files: make(map[string]*File),
	}

	if dirs, ok := d.Dirs[d2.ID]; !ok {
		d.Dirs[d2.ID] = d2
		d.totalItems++
		d.Meta.size += dirs.Meta.size
	}

	d.Meta.Size = ByteCountBinary(d.Meta.size)

	return d2
}

// AddFile Add files to directory
func (d *Dir) AddFile(filename string) *File {
	file := &File{
		ID: uuid.New().String(),
		Meta: Metadata{
			Name:      filename,
			Path:      filepath.Join(d.Meta.Path, filename),
			Timestamp: helpers.MakeTimestampNum(),
		},
	}

	if _, ok := d.Files[file.ID]; !ok {
		if len(d.Files) < 1 {
			d.hasFiles = true
		}
		d.totalItems++
		d.Meta.size += file.Meta.size

		d.Files[file.ID] = file
	}

	return file
}

func (d *Dir) DeleteDir(ids ...string) {
	for _, id := range ids {
		if dir, ok := d.Dirs[id]; ok {
			for _, f := range dir.Files {
				dir.DeleteFile(f.ID)
			}
			for _, dd := range dir.Dirs {
				dd.DeleteDir(dd.ID)
			}
			d.totalItems--
			delete(d.Dirs, dir.ID)
		}
	}
}

func (d *Dir) DeleteFile(ids ...string) {
	for _, id := range ids {
		if f, ok := d.Files[id]; ok {
			d.totalItems--
			d.Meta.size -= f.Meta.size
			delete(d.Files, f.ID)
		}
	}
}

func (d *Dir) Protect(isprotected bool) {
	d.Meta.IsProtected = isprotected
}

func ByteCountBinary(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(b)/float64(div), "KMGTPE"[exp])
}
