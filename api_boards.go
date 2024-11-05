package cvvapi

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"
)

type cvvGetBoardItemResponse struct {
	Id string `json:"id"`
	IdDoc string `json:"id_documento"`
	Type string `json:"tipo_com"`
	TypeDesc string `json:"tipo_com_desc"`
	Title string `json:"titolo"`
	Text string `json:"testo"`
	Chk string `json:"chk"`
	EventDate CvvDate `json:"evento_data"`
	Attachments []struct {
		Id string `json:"id_documento"`
		Title string `json:"descrizione"`
		Chk string `json:"chk"`
	} `json:"allegati"`
} 
type cvvGetBoardItemsResponse struct {
	MsgNew []cvvGetBoardItemResponse `json:"msg_new"`
	MsgRead []cvvGetBoardItemResponse `json:"read"`
}

func boardItemBuilder(item cvvGetBoardItemResponse) BoardItem {
	var res = BoardItem{ 
		Id: item.Id, 
		SingleDoc: item.IdDoc != "",
		Type: item.Type,
		TypeDesc: item.TypeDesc, 
	}

	if item.IdDoc != "" {
		detail := BoardItemDetail{}
		detail.Title = item.Title
		detail.Text = item.Text

		detail.Attachments = []ItemAttachment{
			{
				Id: item.IdDoc,
				Title: item.Title,
				Url: fmt.Sprintf(`https://%v/sdg/app/default/xdownload.php?a=akDOWNLOAD&id=%v&c=%v`, config.cvvHostname, item.IdDoc, item.Chk),
			},
		}

		for _, att := range item.Attachments {
			detail.Attachments = append(detail.Attachments, ItemAttachment{
				Id: att.Id,
				Title: att.Title,
				Url: fmt.Sprintf(`https://%v/sdg/app/default/xdownload.php?a=akDOWNLOAD&id=%v&c=%v`, config.cvvHostname, att.Id, att.Chk),
			})
		}

		res.Detail = &detail
	}

	res.CreatedAt = time.Time(item.EventDate)

	return res
}

// Get all the available board items.
func (o *Session) GetBoardItems(unreadOnly bool) ([]BoardItem, error) {
	var reqUrl = fmt.Sprintf(`https://%v/sif/app/default/bacheca_personale.php`, config.cvvHostname)

	resp, err := o.doPostForm(reqUrl, url.Values{
		"action": []string{"get_comunicazioni"},
	})
	
	if err != nil {
		return nil, err
	}

	var result = new(cvvGetBoardItemsResponse)
	err = resp.getObject(result)

	if err != nil {
		return nil, NewApiError(err)
	}

	
	
	var res []BoardItem

	for _, msg := range result.MsgNew {
		i := boardItemBuilder(msg)
		i.Read = false
		res = append(res, i)
	}

	if !unreadOnly {
		for _, msg := range result.MsgRead {
			i := boardItemBuilder(msg)
			i.Read = true
			res = append(res, i)
		}
	}

	return res, nil
}

// Get a specific board item with its details
func (o *Session) GetBoardItem(boardItemId string) (*BoardItem, error) {

	var boardItem = BoardItem{
		Id: boardItemId,
	}

	err := o.LoadBoardItemDetails(&boardItem)
	if err != nil {
		return nil, err
	}

	return &boardItem, nil
}

// Load the detail of a board item obtained with [GetBoardItems]
func (o *Session) LoadBoardItemDetails(boardItem *BoardItem) error {
	if boardItem.Detail != nil {
		return nil
	}

  var reqUrl = fmt.Sprintf(`https://%v/sif/app/default/bacheca_comunicazione_social.php?com_id=%v`, config.cvvHostname, boardItem.Id);
	
	resp, err := o.doGet(reqUrl)

	if err != nil {
		return err
	}

	var bodyHtml = resp.getHtmlNode()
	var attachmentsHtml = bodyHtml.querySelectorAll(".tabled_layout > div:nth-child(3) a.dwl_allegato")

	detail := BoardItemDetail{}
	detail.Title = strings.TrimSpace(bodyHtml.querySelector(".tabled_layout > div:nth-child(2)").getText())
	if detail.Title == "" {
		detail.Title = "Removed content"
	}

	for _, att := range attachmentsHtml {
		attId := att.getAttr("allegato_id")
		detail.Attachments = append(detail.Attachments, ItemAttachment{
			Id: attId,
			Title: strings.TrimSpace(att.querySelector("div:nth-child(2) > span").getText()),
			Url: fmt.Sprintf(`https://%v/sif/app/default/bacheca_personale.php?action=file_download&com_id=%v`, config.cvvHostname, attId),
		})
	}

	var answerNode *htmlNode

	answerNode = bodyHtml.querySelector(":not(.hidden) > div > .rispondi_bottone")
	if !answerNode.IsEmpty() {
		detail.ExpectConfirmAnswer.AnswerId = answerNode.getAttr("relazione_id")
		detail.ExpectConfirmAnswer.TextHint = strings.TrimSpace((*htmlNode)(answerNode.Parent.Parent).querySelector("div").getText())
	}

	answerNode = bodyHtml.querySelector(":not(.hidden) > div > .rispondi_testo")
	if !answerNode.IsEmpty() {
		detail.ExpectTextAnswer.AnswerId = answerNode.getAttr("relazione_id")
		detail.ExpectTextAnswer.TextHint = strings.TrimSpace((*htmlNode)(answerNode.Parent.Parent).querySelector("div").getText())
		detail.ExpectTextAnswer.Value = bodyHtml.querySelector("#testo_risposta_out").getText()
	}	

	answerNode = bodyHtml.querySelector(":not(.hidden) > div > .rispondi_file")
	if !answerNode.IsEmpty() {
		detail.ExpectAttachmentAnswer.AnswerId = answerNode.getAttr("relazione_id")
		detail.ExpectAttachmentAnswer.TextHint = strings.TrimSpace((*htmlNode)(answerNode.Parent.Parent).querySelector("div").getText())
	}

	boardItem.Detail = &detail

	return nil
}

// Load the attachments files of a board item
func (o *Session) LoadBoardItemAttachments(boardItem *BoardItem) error {
	for idx := range boardItem.Detail.Attachments {
		if boardItem.Detail.Attachments[idx].Resource != nil {
			continue
		}

		res, err := o.GetFileResource(boardItem.Detail.Attachments[idx].Url)
		if err != nil {
			return err
		}
		boardItem.Detail.Attachments[idx].Resource = res
	}

	return nil
}

// Mark a slice of items as read
func (o *Session) MarkBoardItemsAsRead(items []BoardItem) error {
	var reqUrl = fmt.Sprintf(`https://%v/sif/app/default/bacheca_personale.php`, config.cvvHostname)

	var notSingleDocIds = make([]string, 0)
	var singleDocIds = make([]string, 0)
	for _, it := range items {
		if !it.SingleDoc {
			notSingleDocIds = append(notSingleDocIds, it.Id)
		} else {
			singleDocIds = append(singleDocIds, it.Id)
		}
	}

	jsonIds, _ := json.Marshal(notSingleDocIds)

	_, err := o.doPostForm(reqUrl, url.Values{
		"action": []string{"read_all"},
		"id_relazioni": []string{string(jsonIds)},
	})

	if err != nil {
		return err
	}

	for _, id := range singleDocIds {	
		_, err = o.doPostForm(reqUrl, url.Values{
			"action": []string{"lettura_sdg"},
			"com_id": []string{id},
		})
	
		if err != nil {
			return err
		}
	}

	for idx := range items {
		items[idx].Read = true
	}

	return nil
}

// Answer a simple call to action
func (o *Session) ConfirmBoardItem(answerId string) error {
	var reqUrl = fmt.Sprintf(`https://%v/sif/app/default/bacheca_personale.php?action=insert_risposta_bottone&relazione_id=%v`, config.cvvHostname, answerId);

	var _, err = o.doGet(reqUrl)

	if err != nil {
		return err
	}

	return nil
}

// Answer a call to action with text
func (o *Session) AnswerBoardItemWithText(answerId string, text string) error {
	var reqUrl = fmt.Sprintf(`https://%v/sif/app/default/bacheca_personale.php?action=insert_risposta_testo&relazione_id=%v&testo_risposta=%v`, 
		config.cvvHostname, answerId, url.QueryEscape(text));

	var _, err = o.doGet(reqUrl)

	if err != nil {
		return err
	}

	return nil
}

// Answer a call to action with a file
func (o *Session) AnswerBoardItemWithFile(answerId string, file []byte) error {
	//TODO: maybe is formdata not json
	var reqUrl = fmt.Sprintf(`https://%v/sif/app/default/bacheca_personale.php`, config.cvvHostname);

	var payload = fmt.Sprintf(`{
		action: "insert_risposta_file",
		relazione_id: "%v",
		file_risposta: "%v"
	}`, answerId, base64.StdEncoding.EncodeToString(file))

	var err error
	_, err = o.doPostJson(reqUrl, payload)

	if err != nil {
		return err
	}

	return nil
}