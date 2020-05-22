package utils

import (
	"archive/zip"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

// WriteFile will write the info on array of bytes b to filepath. It will set the file
// permission mode to 0660
// Returns an error in case there's any.
func WriteFile(filepath string, b []byte) error {
	return ioutil.WriteFile(filepath, b, 0660)
}

// GetFile will open filepath.
// Returns a tuple with a file and an error in case there's any.
func GetFile(filepath string) (*os.File, error) {
	return os.OpenFile(filepath, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0777)
}

// SafeReadTomlFile will try to acquire a lock on the file and then read its content afterwards.
// Returns an error in case there's any.
func SafeReadTomlFile(filename string, v interface{}) error {
	fileLock := MakeFileMutex(filename)
	fileLock.Lock()
	defer fileLock.Unlock()
	_, err := toml.DecodeFile(filename, v)

	return err
}

// SafeWriteTomlFile will try to acquire a lock on the file and then write to it.
// Returns an error in case there's any.
func SafeWriteTomlFile(v interface{}, filename string) error {
	fileLock := MakeFileMutex(filename)
	fileLock.Lock()
	defer fileLock.Unlock()
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0777)
	defer f.Close()
	if err != nil {
		return err
	}
	encoder := toml.NewEncoder(f)
	return encoder.Encode(v)
}

// DeleteFile will delete filepath permanently.
// Returns an error in case there's any.
func DeleteFile(filepath string) error {
	_, err := os.Stat(filepath)
	if err != nil {
		return err
	}
	err = os.Remove(filepath)
	return err
}

func Unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer rc.Close()

		fpath := filepath.Join(dest, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, f.Mode())
		} else {
			var fdir string
			if lastIndex := strings.LastIndex(fpath, string(os.PathSeparator)); lastIndex > -1 {
				fdir = fpath[:lastIndex]
			}

			err = os.MkdirAll(fdir, f.Mode())
			if err != nil {
				log.Fatal(err)
				return err
			}
			f, err := os.OpenFile(
				fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer f.Close()

			_, err = io.Copy(f, rc)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
