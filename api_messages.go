package cvvapi

import (
	"fmt"
	"log/slog"
	"net/url"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
)

type cvvGetMessagesResponse struct {
	Oas struct {
		Rows []struct {
			Id string `json:"msg_id"`
			//"1" or "0"
			ReadStatus string `json:"read_status"`
			CreatedAt CvvDateTime `json:"dinsert"`
			Sender string `json:"sender"`
			Subject string `json:"oggetto"`
			Body string `json:"testo"`
			HTags []string `json:"htags"`	
		} `json:"rows"`
	} `json:"OAS"`
}

// Get own messages
func (o *Session) GetMessages(unreadOnly bool, page int, pageSize int) ([]Message, error) {
	reqUrl := fmt.Sprintf(`https://%v/sps/app/default/SocMsgApi.php?a=acGetMsgPag`, config.cvvHostname)

	unreadOnlyVal := ifExpr(unreadOnly, 1, 0)

	resp, err := o.doPostForm(reqUrl, url.Values{
		"p": []string{strconv.Itoa(page)},
		"mpp": []string{strconv.Itoa(pageSize)},
		"unreadOnly": []string {strconv.Itoa(unreadOnlyVal)},
	})

	if err != nil {
		return nil, err
	}

	var result = new(cvvGetMessagesResponse)
	err = resp.getObject(result)
	
	if err != nil {
		return nil, NewApiError(err)
	}

	var messages []Message

	for _, msg:= range result.Oas.Rows {
		var item = Message{
			Id: msg.Id,
			Read: msg.ReadStatus == "1",
			CreatedAt: time.Time(msg.CreatedAt),
			Sender: msg.Sender,
			Subject: msg.Subject,
			Body: msg.Body,
			UserLink: fmt.Sprintf(`https://%v%v`, 
				config.cvvHostname, parseHtmlFragment(msg.Body).querySelector("a.msg-userlink").getAttr("href")),
			BoardItemId: "",
			BoardItem: nil,
			ContentItemId: "",
			ContentItem: nil,
		}

		if slices.Index(msg.HTags, "#IFRAMEMESSAGE") >= 0 {
			if slices.Index(msg.HTags, "#BACHECA") >= 0 {
				reg := regexp.MustCompile(`com_id=([^\&]+)&?`)
				item.BoardItemId = reg.FindStringSubmatch(item.UserLink)[1]
			} else if slices.Index(msg.HTags, "#DIDATTICA") >= 0 {
				reg := regexp.MustCompile(`contenuto_id=([^\&]+)&?`)
				item.ContentItemId = reg.FindStringSubmatch(item.UserLink)[1]
			} else {
				slog.Warn("GetMessages: unhandled htags combination", "messageId", item.Id, "htags", msg.HTags)
			}
		}

		messages = append(messages, item)
	}

	return messages, nil
}

// Load message detail (content or boardItem) if any is present
func (o *Session) LoadMessageDetails(message *Message) error {
	var err error
	if message.BoardItemId != "" && message.BoardItem == nil {
		
		if message.BoardItem, err = o.GetBoardItem(message.BoardItemId); err != nil {
			return NewApiError(fmt.Errorf("%v: unable to get board item (messageId=%v, boardItemId=%v)", 
				err.Error(), message.Id, message.BoardItemId))
		}
	}

	if message.ContentItemId != "" && message.ContentItem == nil {
		if message.ContentItem, err = o.GetContentItem(message.ContentItemId); err != nil {
			return NewApiError(fmt.Errorf("%v: unable to get content item (messageId=%v, contentId=%v)", 
				err.Error(), message.Id, message.ContentItemId))
		}
	}

	return nil
}

// Load message attachments files
func (o *Session) LoadMessageAttachments(message *Message) error {
	if message.BoardItem != nil {
		
		if err := o.LoadBoardItemAttachments(message.BoardItem); err != nil {
			return NewApiError(fmt.Errorf("%v: unable to get board item attachments (messageId=%v, boardItemId=%v)", 
				err.Error(), message.Id, message.BoardItemId))
		}
	}

	if message.ContentItem != nil {

		if err := o.LoadContentItemAttachments(message.ContentItem); err != nil {
			return NewApiError(fmt.Errorf("%v: unable to get content item attachments (messageId=%v, contentId=%v)", 
				err.Error(), message.Id, message.ContentItemId))
		}
	}

	return nil
}

// Mark a slice of messages as read
func (o *Session) MarkMessagesAsRead(messages []Message) error {

	var query []string
	for _, item:=range messages {
		query = append(query, fmt.Sprintf("mids[]=%v", item.Id))
	}

	var reqUrl = fmt.Sprintf(`https://%v/sps/app/default/SocMsgApi.php?a=acSetDRead&%v`, config.cvvHostname, strings.Join(query, "&"))

	_, err := o.doGet(reqUrl)

	if err != nil {
		return NewApiError(err)
	}

	for idx := range messages {
		messages[idx].Read = true
	}

	return nil
}