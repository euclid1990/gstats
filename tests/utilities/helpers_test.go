package tests

import (
	"encoding/json"
	"github.com/euclid1990/gstats/utilities"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestExtractPullRequestInfo(t *testing.T) {
	type TypeWant struct {
		Owner  string `json:"owner"`
		Repo   string `json:"repo"`
		Number int    `json:"number"`
		ErrMsg string `json:"err"`
	}
	cases := map[string]struct {
		input string
		want  TypeWant
	}{
		"normal": {
			input: "https://github.com/euclid1990/gstats/pull/1",
			want:  TypeWant{Owner: "euclid1990", Repo: "gstats", Number: 1, ErrMsg: ""},
		},
		"abnormal leak owner": {
			input: "https://github.com/gstats/pull/1",
			want:  TypeWant{Owner: "", Repo: "", Number: 0, ErrMsg: "Can not parse Github pull request link"},
		},
		"abnormal pull id is not number": {
			input: "https://github.com/euclid1990/gstats/pull/abc",
			want:  TypeWant{Owner: "", Repo: "", Number: 0, ErrMsg: "strconv.Atoi: parsing \"abc\": invalid syntax"},
		},
	}
	for tc, td := range cases {
		want, _ := json.Marshal(td.want)
		owner, repo, number, err := utilities.ExtractPullRequestInfo(td.input)
		t.Log(err)
		t.Logf("Execute %s testcase github pull request URL ... (expected %s)\n", tc, string(want))
		assert.Equal(t, td.want.Owner, owner, "Owner should be equal")
		assert.Equal(t, td.want.Repo, repo, "Repo should be equal")
		assert.Equal(t, td.want.Number, number, "Number should be equal")
		if err != nil {
			assert.EqualError(t, err, td.want.ErrMsg, "Err should be equal")
		}
	}
}
