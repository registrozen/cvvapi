package cvvapi_test

import (
	"os"
	"testing"

	"github.com/registrozen/cvvapi"
)

func TestGetStudents(t *testing.T) {
	session, err := cvvapi.NewSession(os.Getenv("CVV_USER"), os.Getenv("CVV_PASSWORD"))

	if err != nil {
		t.Fatal(err)
	}

	courses ,err := session.GetTaughtSubjects()
	
	if err != nil {
		t.Fatal(err)
	}

	if len(courses) == 0 {
		t.Fatalf("to few courses %v", len(courses))
	}
	
	dumpValue(courses, t)

	for _, c := range courses {
		students, err := session.GetStudents(c.ClassId)
		if err != nil {
			t.Fatal(err)
		}
	
		if len(students) == 0 {
			t.Fatalf("to few students %v", len(courses))
		}

		dumpValue(students, t)
	}
}