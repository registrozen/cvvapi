package cvvapi_test

import (
	"os"
	"testing"

	"github.com/registrozen/cvvapi"
)

func TestGetSubjects(t *testing.T) {
	session, err := cvvapi.NewSession(os.Getenv("CVV_USER"), os.Getenv("CVV_PASSWORD"))

	if err != nil {
		t.Fatal(err)
	}

	subjects ,err := session.GetTaughtSubjects()
	
	if err != nil {
		t.Fatal(err)
	}

	if len(subjects) == 0 {
		t.Fatalf("to few courses %v", len(subjects))
	}

	dumpValue(subjects, t)
}