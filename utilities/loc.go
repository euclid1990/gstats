package utilities

import (
	"fmt"
	"github.com/euclid1990/gstats/configs"
	"regexp"
	"strconv"
)

type Loc struct {
	ID      string `json:"id"`
	Name    string `json:"sheet_loc"`
	CGithub string `json:"c_github"`
	CLoc    string `json:"c_loc"`
	Pr      []PR
}

type PR struct {
	Link   string
	Loc    int
	RowNum int
}

func (loc *Loc) getIndexStart() int {
	re := regexp.MustCompile("[0-9]+")
	arrInt := re.FindAllString(loc.CGithub, -1)
	indexStart, err := strconv.Atoi(arrInt[0])
	if err != nil {
		indexStart = 0
	}
	return indexStart
}

func (loc *Loc) ReadLoc(spreadsheet *Spreadsheet) error {
	readRange := fmt.Sprintf("%s!%s:%s", loc.Name, loc.CGithub, loc.CLoc)

	data, err := spreadsheet.read(loc.ID, readRange)
	if err != nil {
		return err
	}

	indexStart := loc.getIndexStart()

	// Receive data that you need
	var pullRequest []PR
	for i, row := range data {
		rowNum := indexStart + i
		newPr := PR{
			RowNum: rowNum,
		}
		if len(row) > 0 {
			if row[0].(string) == configs.GITHUB_TITLE && row[len(row)-1].(string) == configs.LOC_TITLE {
				continue
			}
			lineOfCode, ok := strconv.Atoi(row[len(row)-1].(string))
			if ok != nil {
				lineOfCode = 0
			}
			newPr.Link = row[0].(string)
			newPr.Loc = lineOfCode
		}
		pullRequest = append(pullRequest, newPr)
	}
	loc.Pr = pullRequest
	return nil
}

func (loc *Loc) WriteLoc(spreadsheet *Spreadsheet) error {
	for _, pr := range loc.Pr {
		writeRange := fmt.Sprintf("%s!%s", loc.Name, loc.CLoc+strconv.Itoa(pr.RowNum))
		data := [][]interface{}{
			{pr.Loc},
		}
		err := spreadsheet.write(loc.ID, writeRange, data)
		if err != nil {
			return err
		}
	}
	return nil
}
