package cvvapi_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/registrozen/cvvapi"
)

func init(){
	cvvapi.Config().Use(cvvapi.HttpSnapshoterBuilder("./dev_cache", false))
}

const testing_verbose = false

func dumpValueF(value any, t *testing.T) {
	buffer := bytes.Buffer{}
	encoder := json.NewEncoder(&buffer)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	err := encoder.Encode(value)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(buffer.String())
}

func dumpValue(value any, t *testing.T) {
	if !testing_verbose {
		return
	}

	dumpValueF(value, t)
}