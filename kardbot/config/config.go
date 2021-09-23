package config

// TODO: Make this a go library with its own repo? This has potential to be useful for
//       other projects. It would be nice to aggregate all provided config, possibly
//       multiple files, of a given format (e.g., JSON) as a single byte slice.
//       Could also provide helper methods to populate structs with config data.

import (
	"fmt"
	"io/ioutil"
	"os"

	log "github.com/sirupsen/logrus"
)

// Returns the raw bytes contained in jsonConfigFile
// so that other components can unmarshal it as desired.
var RawJSONConfig func() []byte

// The config file is expected to live at under
// projectRoot/config
const configFilename = "config.json"

func init() {
	filepath := fmt.Sprintf("config/%s", configFilename)

	fd, err := os.Open(filepath)
	if err != nil {
		log.Fatal(err)
	}
	defer fd.Close()

	rawJSON, err := ioutil.ReadAll(fd)
	if err != nil {
		log.Fatal(err)
	}

	RawJSONConfig = func() []byte { return rawJSON }
}
