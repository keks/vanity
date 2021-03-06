package vanity

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var addr = "https://kkn.fi"

func TestRedirectFromHttpToHttps(t *testing.T) {
	res := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "http://kkn.fi", nil)
	if err != nil {
		t.Fatal(err)
	}
	srv := Redirect("git", "kkn.fi", "https://github.com/kare")
	srv.ServeHTTP(res, req)
	if res.Code != http.StatusMovedPermanently {
		t.Fatalf("expected response status 301, but got %v", res.Code)
	}
	if res.Header().Get("Location") != addr {
		t.Fatalf("expected response location '%v', but got '%v'", addr, res.Header().Get("Location"))
	}
}

func TestHTTPMethodsSupport(t *testing.T) {
	tests := []struct {
		method string
		status int
	}{
		{http.MethodGet, http.StatusOK},
		{http.MethodHead, http.StatusMethodNotAllowed},
		{http.MethodPost, http.StatusMethodNotAllowed},
		{http.MethodPut, http.StatusMethodNotAllowed},
		{http.MethodDelete, http.StatusMethodNotAllowed},
		{http.MethodTrace, http.StatusMethodNotAllowed},
		{http.MethodOptions, http.StatusMethodNotAllowed},
	}
	for _, test := range tests {
		req, err := http.NewRequest(test.method, addr+"/gist?go-get=1", nil)
		if err != nil {
			t.Skipf("http request with method %v failed with error: %v", test.method, err)
		}
		res := httptest.NewRecorder()
		srv := Redirect("git", "kkn.fi", "https://github.com/kare")
		srv.ServeHTTP(res, req)
		if res.Code != test.status {
			t.Fatalf("Expecting status code %v for method '%v', but got %v", test.status, test.method, res.Code)
		}
	}
}

func TestIndexPageNotFound(t *testing.T) {
	res := httptest.NewRecorder()
	req, err := http.NewRequest("GET", addr, nil)
	if err != nil {
		t.Fatal(err)
	}
	srv := Redirect("git", "go.cryptoscope.co/vanity", "https://github.com/keks/vanity")
	srv.ServeHTTP(res, req)
	if res.Code != http.StatusNotFound {
		t.Fatalf("Expected response status 404, but got %v", res.Code)
	}
}

func TestGoTool(t *testing.T) {
	tests := []struct {
		path   string
		result string
	}{
		{"/vanity/?go-get=1", "go.cryptoscope.co/vanity git https://github.com/keks/vanity"},
		{"/vanity/cmd/?go-get=1", "go.cryptoscope.co/vanity git https://github.com/keks/vanity"},
		{"/vanity/cmd/vanity?go-get=1", "go.cryptoscope.co/vanity git https://github.com/keks/vanity"},
		{"/vanity/doesnt-even-exist?go-get=1", "go.cryptoscope.co/vanity git https://github.com/keks/vanity"},
	}
	for _, test := range tests {
		res := httptest.NewRecorder()
		req, err := http.NewRequest("GET", addr+test.path, nil)
		if err != nil {
			t.Fatal(err)
		}
		req.Host="go.cryptoscope.co"
		srv := Redirect("git", "go.cryptoscope.co/vanity", "https://github.com/keks/vanity")
		srv.ServeHTTP(res, req)

		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("reading response body failed with error: %v", err)
		}

		expected := `<meta name="go-import" content="` + test.result + `">`
		if !strings.Contains(string(body), expected) {
			t.Fatalf("Expecting url '%v' body to contain html meta tag: '%v', but got:\n'%v'", test.path, expected, string(body))
		}

		expected = "text/html; charset=utf-8"
		if res.HeaderMap.Get("content-type") != expected {
			t.Fatalf("Expecting content type '%v', but got '%v'", expected, res.HeaderMap.Get("content-type"))
		}

		if res.Code != http.StatusOK {
			t.Fatalf("Expected response status 200, but got %v", res.Code)
		}
	}
}

func TestBrowserGoDoc(t *testing.T) {
	tests := []struct {
		path   string
		result string
	}{
		{"/gist", "https://godoc.org/kkn.fi/gist"},
		{"/set", "https://godoc.org/kkn.fi/set"},
		{"/cmd/vanity", "https://godoc.org/kkn.fi/cmd/vanity"},
		{"/cmd/tcpproxy", "https://godoc.org/kkn.fi/cmd/tcpproxy"},
		{"/pkgabc/sub/foo", "https://godoc.org/kkn.fi/pkgabc/sub"},
	}
	for _, test := range tests {
		res := httptest.NewRecorder()
		req, err := http.NewRequest("GET", addr+test.path, nil)
		if err != nil {
			t.Fatal(err)
		}
		srv := Redirect("git", "kkn.fi", "https://github.com/kare")
		srv.ServeHTTP(res, req)

		if res.Code != http.StatusTemporaryRedirect {
			t.Fatalf("Expected response status %v, but got %v", http.StatusTemporaryRedirect, res.Code)
		}
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("reading response body failed with error: %v", err)
		}
		if !strings.Contains(string(body), test.result) {
			t.Fatalf("Expecting '%v' be contained in '%v'", test.result, string(body))
		}
	}
}
