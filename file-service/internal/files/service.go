package files

import (
	"errors"
	"io/fs"
	"os"
)

const (
	FilesDir = "uploads/"
)

type Service struct {
}

func NewService() (*Service, error) {
	_, err := os.Stat(FilesDir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			err = os.Mkdir(FilesDir, 0777)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	return &Service{}, nil
}

func (s *Service) SetFile(filename string, bytes []byte) error {
	return os.WriteFile(FilesDir+filename, bytes, 0666)
}

func (s *Service) DeleteFile(filename string) error {

	_, err := os.Stat(FilesDir + filename)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return err
	}

	return os.Remove(FilesDir + filename)
}

func (s *Service) GetFilePath(name string) string {
	return FilesDir + name
}

func (s *Service) FileExists(filename string) bool {
	_, err := os.Stat(FilesDir + filename)
	if err != nil {
		return false
	}
	return true
}
