package kardbot

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
)

// Returns the raw bytes contained in jsonConfigFile
// so that other components can unmarshal it as desired.
var rawJSONConfig func() []byte

// The config file is expected to live at under
// projectRoot/config
const configFilename = "config.json"

func init() {
	_, b, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatal("Could not retrieve project root")
	}
	basepath := filepath.Dir(b)

	filepath := fmt.Sprintf("%s/../config/%s", basepath, configFilename)

	fd, err := os.Open(filepath)
	if err != nil {
		log.Fatal(err)
	}
	defer fd.Close()

	rawJSON, err := ioutil.ReadAll(fd)
	if err != nil {
		log.Fatal(err)
	}

	rawJSONConfig = func() []byte { return rawJSON }
}
