package main

import (
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.cryptoscope.co/vanity"
)

func TestImport(t *testing.T) {
	type testcase struct {
		imprt    *vanity.Repo
		srvHost  string
		reqPaths []string
		status   []int
	}

	test := func(tc testcase) func(*testing.T) {
		return func(t *testing.T) {
			a := assert.New(t)

			for i, reqPath := range tc.reqPaths {
				rr := httptest.NewRecorder()

				req := httptest.NewRequest("GET", reqPath, nil)
				req.Header.Set("Host", tc.srvHost)

				tc.imprt.ServeHTTP(rr, req)

				a.Equal(tc.status[i], rr.Code, "status code mismatch for %v", i)
				t.Logf("%#v",rr)
			}
		}
	}

	tcs := []testcase{
		{
			imprt: &vanity.Repo{
				ImportPrefix:     "go.cryptoscope.co/vanityd",
				VCS:      "git",
				RepoRoot: "github.com/keks/vanityd",
			},
			srvHost: "go.cryptoscope.co",
			reqPaths: []string{
				"http://go.cryptoscope.co/vanityd",
				"https://go.cryptoscope.co/vanityd/foo",
				"https://go.cryptoscope.co/vanityd/foo?go-get=1",
				"https://go.cryptoscope.co/vanityd",
				"https://go.cryptoscope.co/vanityd?go-get=1",
			},
			status: []int{
				301,
				307,
				200,
				307,
				200,
			},
		},
	}

	for i, tc := range tcs {
		t.Run(fmt.Sprint(i), test(tc))
	}
}
