package cvvapi

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Get students of a given class
func (o *Session) GetClasses() ([]Class, error) {
	var reqUrl = fmt.Sprintf(`https://%v/cvv/app/default/selezione_classi.php`, config.cvvHostname)

	resp, err := o.doGet(reqUrl)

	if err != nil {
		return nil, err
	}

	var bodyNode = resp.getHtmlNode()
	var rowNodes = bodyNode.querySelectorAll(`tr[height="38"]`)

	var regClassId = regexp.MustCompile(`regclasse.php\?classe_id=(.+)&`)
	
	reqUrl = fmt.Sprintf(`https://%v/cvv/app/default/gioprof_selezione.php`, config.cvvHostname)

	resp, err = o.doGet(reqUrl)

	if err != nil {
		return nil, err
	}

	var ownBodyNode = resp.getHtmlNode()
	var ownedSections = mapFunc(ownBodyNode.querySelectorAll(
		"table.griglia_tab:nth-child(2) > tbody > tr[valign=top] td:nth-child(1) > a"), 
		func(n *htmlNode) string { return strings.TrimSpace(n.getText()) })

	var result = []Class{}

	for _, rn := range rowNodes[1:] {

		item := Class{}

		item.Section = strings.TrimSpace(rn.querySelector("td:nth-child(1) div").getText())
		item.Building = strings.Fields(rn.querySelector("td:nth-child(2) p:nth-child(2)").getText())[1]

		node := rn.querySelector("td div.rigtab a")

		item.Name = strings.TrimSpace(node.querySelector("div:nth-child(1)").getText())
		grade, _ := strconv.Atoi(item.Name[0:1])
		item.Grade = grade
		item.Course = strings.TrimSpace(node.querySelector("div:nth-child(2)").getText())
		
		link := 	node.getAttr("href")
		item.Id = regClassId.FindStringSubmatch(link)[1]
		item.Link = fmt.Sprintf(`https://%v/cvv/app/default/%v`, config.cvvHostname, link)

		item.Owned = findFirst(ownedSections, func (p *string) bool { return item.Name == *p }) != nil

		result = append(result, item)
	}

	return result, nil
}