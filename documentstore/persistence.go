package documentstore

import (
	"encoding/gob"
	"encoding/json"
	"encoding/xml"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"

	"github.com/vaiktorg/grimoire/helpers"
)

// DirToFS Creates a directory in the document store
func (d *Dir) DirToFS() error {
	if !d.Meta.IsProtected {
		err := os.MkdirAll(d.Meta.Path, os.ModePerm)
		if err != nil {
			return err
		}
	}
	return nil
}

func (f *File) FileToFS() error {
	if !f.Meta.IsProtected {
		file, err := os.Create(f.Meta.Path)
		if err != nil {
			return err
		}
		defer file.Close()

		if f.Data != nil && len(f.Data) > 0 {
			_, err = file.Write(f.Data)
			if err != nil {
				return err
			}
			f.Meta.size = int64(cap(f.Data))
			f.Meta.Size = strconv.Itoa(int(f.Meta.size))
		}
	}
	return nil
}

// CreateStructOnFS creates structure set on document store
func (d *Dir) CreateStructOnFS() error {

	err := d.DirToFS()
	if err != nil {
		return err
	}

	for _, file := range d.Files {
		err := file.FileToFS()
		if err != nil {
			return err
		}
	}

	for _, item := range d.Dirs {
		err := item.CreateStructOnFS()
		if err != nil {
			return err
		}

	}

	return nil
}

type Format string

const (
	JSON Format = "json"
	YAML        = "yaml"
	GOB         = "gob"
	XML         = "xml"
)

// Load loads json file to populate directory
func (d *Dir) Load(fmt Format) error {
	file, err := helpers.OpenFile("FSState_" + helpers.MakeTimestampNum() + "." + string(fmt))
	defer file.Close()

	if err != nil {
		return err
	}

	switch fmt {
	case XML:
		err = xml.NewDecoder(file).Decode(d)
	case GOB:
		err = gob.NewDecoder(file).Decode(d)
	case YAML:
		err = yaml.NewDecoder(file).Decode(d)
	case JSON:
		err = json.NewDecoder(file).Decode(d)
	}

	if err != nil {
		return err
	}
	return nil
}

// Persist directory for later reference and persistence.
func (d *Dir) Persist(fmt Format) error {
	file, err := helpers.OpenFile("FSState_" + helpers.MakeTimestampNum() + "." + string(fmt))
	defer file.Close()

	if err != nil {
		return err
	}

	switch fmt {
	case XML:
		err = xml.NewEncoder(file).Encode(file)
	case GOB:
		err = gob.NewEncoder(file).Encode(file)
	case YAML:
		err = yaml.NewEncoder(file).Encode(file)
	case JSON:
		err = json.NewEncoder(file).Encode(d)
	}

	if err != nil {
		return err
	}
	return nil
}
