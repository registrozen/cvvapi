package cvvapi_test

import (
	"os"
	"testing"

	"github.com/registrozen/cvvapi"
)

func TestGetClasses(t *testing.T) {
	session, err := cvvapi.NewSession(os.Getenv("CVV_USER"), os.Getenv("CVV_PASSWORD"))

	if err != nil {
		t.Fatal(err)
	}

	classes ,err := session.GetClasses()
	
	if err != nil {
		t.Fatal(err)
	}

	if len(classes) == 0 {
		t.Fatalf("to few classes %v", len(classes))
	}

	dumpValue(classes, t)
}