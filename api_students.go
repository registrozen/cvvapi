package cvvapi

import (
	"fmt"
	"strings"
)

// Get students of a given class
func (o *Session) GetStudents(classId string) ([]Student, error) {
	var reqUrl = fmt.Sprintf(`https://%v/cvv/app/default/regclasse.php?classe_id=%v&gruppo_id=`, config.cvvHostname, classId)

	resp, err := o.doGet(reqUrl)

	if err != nil {
		return nil, err
	}

	var bodyNode = resp.getHtmlNode()
	var rowNodes = bodyNode.querySelectorAll("table#data_table_2 > tbody > tr.rigtab")

	var result = []Student{}

	for _, rn := range rowNodes[1:] {

		names := rn.querySelector("td:nth-child(4) > div:nth-child(1)")

		fullName := strings.TrimSpace(names.getText())
		parts := strings.Split(fullName, " ")
	
		item := Student{
			FullName: fullName,
		}
		
		(&item.BirthDay).UnmarshalJSON([]byte(strings.Fields(rn.querySelector("td:nth-child(4) > div:nth-child(2)").getText())[0]))

		if len(parts) < 3 {
			item.Names = parts[1:]
			item.Surnames = parts[0:1]
		} else {
			item.Names = parts[2:]
			item.Surnames = parts[0:2]
		}

		status := strings.TrimSpace(rn.querySelector("td:nth-child(7) p:nth-child(1)").getText())
		switch strings.ToUpper(status) {
		case "P":
			item.Status = "present"
		case "A":
			item.Status = "absent"
		case "R":
			item.Status = "late"
		case "U":
			item.Status = "exited"
		case "XG":
			item.Status = "nolesson"
		default:
			fmt.Printf("%v, %v", status, status == "P")
			item.Status = "unknown"
		}

		result = append(result, item)
	}

	return result, nil
}