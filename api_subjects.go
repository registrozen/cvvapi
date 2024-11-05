package cvvapi

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Get own courses
func (o *Session) GetTaughtSubjects() ([]Subject, error) {
	var reqUrl = fmt.Sprintf(`https://%v/cvv/app/default/gioprof_selezione.php`, config.cvvHostname)

	resp, err := o.doGet(reqUrl)

	if err != nil {
		return nil, err
	}

	var regClassId = regexp.MustCompile(`regclasse.php\?classe_id=(.+)&`)
	var regSubjectId = regexp.MustCompile(`materia=([^&]+)&`)
	var bodyNode = resp.getHtmlNode()
	var rowNodes = bodyNode.querySelectorAll("table.griglia_tab:nth-child(2) > tbody > tr[valign=top]")

	var result = []Subject{}

	for _, rn := range rowNodes[1:] {
		section := 	strings.TrimSpace(rn.querySelector("td:nth-child(1) > a").getText())
		link := 	strings.TrimSpace(rn.querySelector("td:nth-child(1) > a").getAttr("href"))
		subjectNodes := rn.querySelectorAll("td:nth-child(3) > div")
		classId := regClassId.FindStringSubmatch(link)[1]

		for _, subjectNode := range subjectNodes {
			grade, _ := strconv.Atoi(section[0:1])
			result = append(result, Subject{
				ClassSection: section[1:],
				Name: strings.TrimSpace(subjectNode.querySelector("div > div.open_sans_condensed_bold").getAttr("title")),
				Id: regSubjectId.FindStringSubmatch(
					strings.TrimSpace(subjectNode.querySelector("div > a").getAttr("href")))[1],
				ClassLink: fmt.Sprintf(`https://%v/cvv/app/default/%v`, config.cvvHostname, link),
				ClassId: classId,
				ClassGrade: grade,
				ClassName: section,
			})
		}
	}

	return result, nil
}