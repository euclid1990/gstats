package utilities

import (
	"fmt"
	"strconv"
)

const REDMINE_URL_SPLIT = "/issues/"

type Loc struct {
	ID        string `json:"id"`
	Name      string `json:"sheet_loc"`
	CGithub   string `json:"c_github"`
	CTicket   string `json:"c_ticket"`
	CPoint    string `json:"c_point"`
	CLoc      string `json:"c_loc"`
	CRowStart int    `json:"c_row_start"`
	Pr        []PR
	UpdatedPr int
	SheetId   string
}

type PR struct {
	Link     string
	IDTicket int
	Point    int
	Loc      int
	RowNum   int
}

func (loc *Loc) ReadLoc(spreadsheet *Spreadsheet) error {
	loc.SheetId = spreadsheet.GetGidBySheetName(loc.ID, loc.Name)

	minCol, maxCol := GetMinMaxCharacter(loc.CTicket, loc.CGithub, loc.CLoc, loc.CPoint)

	indexGithub := GetColumnDistance(minCol, loc.CGithub)
	indexTicket := GetColumnDistance(minCol, loc.CTicket)

	readRange := fmt.Sprintf("%s!%s:%s", loc.Name, minCol+strconv.Itoa(loc.CRowStart), maxCol)

	data, err := spreadsheet.read(loc.ID, readRange)
	if err != nil {
		return err
	}

	// Receive data that you need
	var pullRequest []PR
	for i, row := range data {
		rowNum := loc.CRowStart + i
		newPr := PR{
			RowNum: rowNum,
		}
		rowLength := len(row)
		if rowLength > 0 {
			// Get Github Link
			if indexGithub < rowLength {
				newPr.Link = row[indexGithub].(string)
			}

			// Get ID Ticket
			if indexTicket < rowLength {
				ticket := row[indexTicket].(string)
				newPr.IDTicket = GetIDTicket(ticket, REDMINE_URL_SPLIT)
			}

		}
		pullRequest = append(pullRequest, newPr)
	}

	loc.Pr = pullRequest
	return nil
}

func (loc *Loc) WriteLoc(spreadsheet *Spreadsheet) error {
	minCol, maxCol := GetMinMaxCharacter(loc.CLoc, loc.CPoint)

	indexPoint := GetColumnDistance(minCol, loc.CPoint)
	indexLoc := GetColumnDistance(minCol, loc.CLoc)

	writeLength := GetColumnDistance(minCol, maxCol) + 1

	for _, pr := range loc.Pr {
		// Write Loc
		rowNum := strconv.Itoa(pr.RowNum)
		writeRange := fmt.Sprintf("%s!%s:%s", loc.Name, loc.CPoint+rowNum, loc.CLoc+rowNum)

		values := make([]interface{}, writeLength)
		values[indexPoint] = pr.Point
		values[indexLoc] = pr.Loc

		data := [][]interface{}{
			values,
		}

		err := spreadsheet.write(loc.ID, writeRange, data)
		if err != nil {
			return err
		}
		loc.UpdatedPr++
	}

	return nil
}
