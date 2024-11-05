package cvvapi

import (
	"encoding/json"
	"strings"
	"time"
)

// Date in ClasseViva format
type CvvDate time.Time

func (o *CvvDate) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	t, err := time.Parse("02-01-2006", s)
	if err != nil {
		return err
	}
	*o = CvvDate(t)
	return nil
}
	
func (o CvvDate) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Time(o))
}

// Date and time in ClasseViva format
type CvvDateTime time.Time

func (o *CvvDateTime) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	t, err := time.Parse(time.DateTime, s)
	if err != nil {
		return err
	}
	*o = CvvDateTime(t)
	return nil
}
	
func (o CvvDateTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Time(o))
}

// Resource obtained as a download from the download resource endpoint
type FileResource struct {
	Name string
	ContentType string
	Data []byte
}

// Attachment descriptor, both content and board
type ItemAttachment struct {
	Id   string
	Title string
	Url  string
	Link string
	Resource *FileResource
}

// Message content item descriptor
type ContentItem struct {
	Id       string
	Title    string
	FolderId string
	Attachments []ItemAttachment
}

// Board item descriptor
type BoardItem struct {
	Id string
	CreatedAt time.Time
	Type string
	TypeDesc string
	Read bool
	SingleDoc bool
	Detail *BoardItemDetail
}

// Board item detail descriptor
type BoardItemDetail struct {
	Title string
	Text string
	Attachments []ItemAttachment
	ExpectConfirmAnswer BoardItemAnswer
	ExpectTextAnswer BoardItemTextAnswer
	ExpectAttachmentAnswer BoardItemAnswer
}

// Board item answer descriptor
type BoardItemAnswer struct {
	AnswerId string
	TextHint string
}

// Board item text answer descriptor
type BoardItemTextAnswer struct {
	BoardItemAnswer
	Value string
}

// Check if an answer is expected for a board item
func (o *BoardItem) ExpectAnswer() bool {
	return o != nil && o.Detail != nil && (
		o.Detail.ExpectConfirmAnswer.AnswerId != "" || 
		o.Detail.ExpectTextAnswer.AnswerId != "" || 
		o.Detail.ExpectAttachmentAnswer.AnswerId != "" )
}

// Message descriptor
type Message struct {
	Id string
	Read bool
	CreatedAt time.Time
	Sender string
	Subject string
	Body string
	UserLink string
	BoardItemId string
	BoardItem *BoardItem
	ContentItemId string
	ContentItem *ContentItem
}

// Subject descriptor
type Subject struct {
	Id string
	Name string
	ClassId string
	ClassGrade int
	ClassName string
	ClassSection string
	ClassLink string
}

// Class descriptor
type Class struct {
	Id string
	Name string
	Grade int
	Section string
	Building string
	Course string
	Owned bool
	Link string
}

// Student descriptor
type Student struct {
	FullName string
	Names []string
	Surnames []string
	BirthDay CvvDate
	Status string // present, absent, late, exited, nolesson
}