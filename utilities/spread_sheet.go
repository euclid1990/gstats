package utilities

import (
	"encoding/json"
	"github.com/euclid1990/gstats/configs"
	"golang.org/x/sync/errgroup"
	"google.golang.org/api/sheets/v4"
	"io/ioutil"
	"net/http"
)

const SPREADSHEET_VALUE_INPUT_RAW = "RAW"

type Spreadsheet struct {
	srv *sheets.Service
	err error
}

func NewSheet(client *http.Client) *Spreadsheet {
	srv, err := sheets.New(client)
	spreadsheet := &Spreadsheet{
		srv: srv,
		err: err,
	}
	return spreadsheet
}

func getSpreadSheets() ([]Loc, error) {
	raw, err := ioutil.ReadFile(configs.SPREADSHEET_JSON)
	if err != nil {
		return nil, err
	}

	var data []Loc
	err = json.Unmarshal(raw, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (spreadsheet *Spreadsheet) read(spreadsheetId, readRange string) ([][]interface{}, error) {
	// readRange: Range will be read data - Type: string - Format: "Sheetname!AddressStart:ColumnEnd" - Example: "LOC!A6:I"
	if spreadsheet.err != nil {
		return nil, spreadsheet.err
	}
	srv := spreadsheet.srv
	resp, err := srv.Spreadsheets.Values.Get(spreadsheetId, readRange).Do()
	if err != nil {
		return nil, err
	}

	if len(resp.Values) > 0 {
		return resp.Values, nil
	}
	return nil, nil
}

func (spreadsheet *Spreadsheet) write(spreadsheetId string, writeRange string, data [][]interface{}) error {
	// writeRange: Range will be written data - Type: string - Format: "Sheetname!AddressStart:ColumnEnd" - Example: "LOC!A3:C"
	if spreadsheet.err != nil {
		return spreadsheet.err
	}
	srv := spreadsheet.srv
	var vr sheets.ValueRange
	vr.Values = data

	_, err := srv.Spreadsheets.Values.Update(spreadsheetId, writeRange, &vr).ValueInputOption(SPREADSHEET_VALUE_INPUT_RAW).Do()
	if err != nil {
		return err
	}
	return nil
}

func (spreadSheet *Spreadsheet) UpdateLocSpreadsheets() error {
	eg := errgroup.Group{}
	sheets, err := getSpreadSheets()
	if err != nil {
		return err
	}

	for i, _ := range sheets {
		sh := &sheets[i]
		eg.Go(func() error {
			err := sh.ReadLoc(spreadSheet)
			if err != nil {
				return err
			}

			// github.UpdateAddtionCode(sh)

			err = sh.WriteLoc(spreadSheet)
			if err != nil {
				return err
			}
			return nil
		})
	}
	defer SendLocMessage(sheets)

	if err = eg.Wait(); err != nil {
		return err
	}

	return nil
}
