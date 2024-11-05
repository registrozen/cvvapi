package cvvapi_test

import (
	"os"
	"slices"
	"testing"

	"github.com/registrozen/cvvapi"
)

func TestGetBoardItems(t *testing.T) {
	session, err := cvvapi.NewSession(os.Getenv("CVV_USER"), os.Getenv("CVV_PASSWORD"))

	if err != nil {
		t.Fatal(err)
	}

	items, err := session.GetBoardItems(false)
	
	if err != nil {
		t.Fatal(err)
	}

	if len(items) == 0 {
		t.Fatalf("to few items %v", len(items))
	}

	dumpValue(items[0], t)
}

func TestLoadBoardItemDetails(t *testing.T) {
	session, err := cvvapi.NewSession(os.Getenv("CVV_USER"), os.Getenv("CVV_PASSWORD"))

	if err != nil {
		t.Fatal(err)
	}

	items, err := session.GetBoardItems(false)
	
	if err != nil {
		t.Fatal(err)
	}

	if len(items) == 0 {
		t.Fatalf("to few items %v", len(items))
	}	
	
	idx := slices.IndexFunc(items, func (m cvvapi.BoardItem) bool {
		return m.Id == "909723"
	})

	err = session.LoadBoardItemDetails(&items[idx])

	if err != nil {
		t.Fatal(err)
	}

	if items[idx].Detail.Title == "" {
		t.Fatal("item not loaded")
	}

	dumpValue(items[idx], t)
}

func TestDownloadAttachment(t *testing.T) {
	session, err := cvvapi.NewSession(os.Getenv("CVV_USER"), os.Getenv("CVV_PASSWORD"))

	if err != nil {
		t.Fatal(err)
	}

	items, err := session.GetBoardItems(false)
	
	if err != nil {
		t.Fatal(err)
	}

	if len(items) == 0 {
		t.Fatalf("to few items %v", len(items))
	}	
	
	idx := slices.IndexFunc(items, func (m cvvapi.BoardItem) bool {
		return m.Id == "909723"
	})

	err = session.LoadBoardItemDetails(&items[idx])

	if err != nil {
		t.Fatal(err)
	}

	if items[idx].Detail.Title == "" {
		t.Fatal("item not loaded")
	}

	att, err := session.GetFileResource(items[idx].Detail.Attachments[0].Url)

	if err != nil {
		t.Log(att)
		t.Fatal(err)
	}
	
	// dumpValueF(items[idx].Attachments[0], t)
	// os.WriteFile("./dumpfiles/"+att.Name, att.Data, os.ModePerm)
}