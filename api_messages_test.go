package cvvapi_test

import (
	"os"
	"slices"
	"testing"

	"github.com/registrozen/cvvapi"
)

func TestGetMessages(t *testing.T) {
	session, err := cvvapi.NewSession(os.Getenv("CVV_USER"), os.Getenv("CVV_PASSWORD"))

	if err != nil {
		t.Fatal(err)
	}

	messages, err := session.GetMessages(false, 1, 50)
	
	if err != nil {
		t.Fatal(err)
	}

	if len(messages) == 0 {
		t.Fatalf("to few messages %v", len(messages))
	}

	dumpValue(messages[1], t)
}


func TestLoadMessageDetails(t *testing.T) {
	session, err := cvvapi.NewSession(os.Getenv("CVV_USER"), os.Getenv("CVV_PASSWORD"))

	if err != nil {
		t.Fatal(err)
	}

	messages, err := session.GetMessages(false, 1, 50)
	
	if err != nil {
		t.Fatal(err)
	}

	if len(messages) == 0 {
		t.Fatalf("to few messages %v", len(messages))
	}

	bIdx := slices.IndexFunc(messages, func (m cvvapi.Message) bool {
		return m.BoardItemId != ""
	})
	
	// cIdx := slices.IndexFunc(messages, func (m cvvapi.CvvMessage) bool {
	// 	return m.ContentItemId != ""
	// })

	err = session.LoadMessageDetails(&messages[bIdx])

	if err != nil {
		t.Fatal(err)
	}
	dumpValue(messages[bIdx], t)

	// //TODO: find message with contentId
	// err = session.LoadMessageDetails(&messages[cIdx])

	// if err != nil {
	// 	t.Fatal(err)
	// }
}