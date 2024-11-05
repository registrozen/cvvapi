package cvvapi

import (
	"fmt"
	"io"
	"log/slog"
	"regexp"
	"strings"
)

// Get a content item of a message
func (o *Session) GetContentItem(contentId string) (*ContentItem, error) {
	reqUrl := fmt.Sprintf(`https://%v/cvv/app/default/didattica_social_view.php?contenuto_id=%v`, config.cvvHostname, contentId)

	resp, err := o.doGet(reqUrl)

	if err != nil {
		return nil, err
	}

	var bodyNode = resp.getHtmlNode()
	
	var attachments = bodyNode.querySelectorAll("tr.row.contenuto")

	var result = &ContentItem{
		Id: contentId, 
		Title: strings.TrimSpace(bodyNode.querySelector("td[folder_id]").getHtml()),
		FolderId: bodyNode.querySelector("td[folder_id]").getAttr("folder_id"),
	}

	result.Attachments = []ItemAttachment{}
	for _, att := range attachments {
		item := ItemAttachment{
			Id: att.getAttr("contenuto_id"),
			Title: strings.TrimSpace(att.querySelector(".contenuto_desc span:nth-child(1)").getText()),
			Url: "",
			Link: "",
		}

		if val := strings.TrimSpace(att.querySelector(".button_action.action_download").getAttr("cksum")); val != "" {
			item.Url = fmt.Sprintf(
				`https://%v/fml/app/default/didattica_genitori.php?a=downloadContenuto&contenuto_id=%v&cksum=%v`, 
				config.cvvHostname, item.Id, val)
		} else if val := strings.TrimSpace(att.querySelector(".button_action.action_link").getAttr("ref")); val != "" {
			item.Link = val
		} else {
			slog.Warn("GetContentItem: Unknown content resource", "contentId", contentId, "attId", item.Id)
		}	
		result.Attachments = append(result.Attachments, item)	
	}
	

	return result, nil
}


func (o *Session) LoadContentItemAttachments(contentItem *ContentItem) error {
	for idx := range contentItem.Attachments {
		if contentItem.Attachments[idx].Resource != nil {
			continue
		}

		res, err := o.GetFileResource(contentItem.Attachments[idx].Url)
		if err != nil {
			return err
		}
		contentItem.Attachments[idx].Resource = res
	}

	return nil
}

func (o *Session) GetFileResource(resourceUrl string) (*FileResource, error) {
	resp, err := o.doGet(resourceUrl)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	var reg = regexp.MustCompile("attachment; filename=(.+)$")

	var result = &FileResource{
		Name: reg.FindStringSubmatch(resp.Header.Get("Content-Disposition"))[1],
		ContentType: resp.Header.Get("Content-Type"),
		Data: data,
	}

	return result, nil
}