package config

import (
	"io/ioutil"
	"os"
)

type jsonConfig struct {
	Raw      []byte
	filepath string
}

func NewJsonConfig(filepath string) (*jsonConfig, error) {
	j := jsonConfig{filepath: filepath}

	fd, err := os.Open(j.filepath)
	if err != nil {
		return nil, err
	}
	defer fd.Close()

	raw, err := ioutil.ReadAll(fd)
	if err != nil {
		return nil, err
	}
	j.Raw = raw

	return &j, nil
}

func (j *jsonConfig) Filepath() string {
	return j.filepath
}
