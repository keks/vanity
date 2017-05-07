package main // import "kkn.fi/cmd/vanity"

import (
	"fmt"
	"net/http"
	"strings"
)

type (
	// packageConfig defines Go package that has vanity import defined by Path,
	// VCS system type and VCS URL.
	packageConfig struct {
		// Name is the name of the Go package.
		Name string
		// VCS is version control system used by the project.
		VCS string
		// URL of the git repository
		URL string
	}
	// vanityServer is the actual HTTP server for Go vanity domains.
	vanityServer struct {
		// Domain is the vanity domain.
		Domain string
		// Packages contains settings for vanity Packages.
		Packages map[string]*packageConfig
	}
)

func (p packageConfig) name() string {
	path := p.Name
	c := strings.Index(path, "/")
	if c == -1 {
		return path
	}
	return path[c+1:]
}

// newPackage returns a new Package given a path and VCS.
func newPackage(name, vcs, url string) *packageConfig {
	return &packageConfig{
		Name: name,
		VCS:  vcs,
		URL:  url,
	}
}

// newServer returns a new Vanity Server given domain name and
// vanity package configuration.
func newServer(domain string, config map[string]*packageConfig) *vanityServer {
	return &vanityServer{
		Domain:   domain,
		Packages: config,
	}
}

// goMetaContent creates a value from the <meta/> tag content attribute.
func (p packageConfig) goMetaContent() string {
	return fmt.Sprintf("%v %v", p.VCS, p.URL)
}

// goDocURL returns the HTTP URL to godoc.org.
func (p packageConfig) goDocURL(domain string) string {
	return fmt.Sprintf("https://godoc.org/%v%v", domain, p.Name)
}

// goImportLink creates the link used in HTML <meta/> tag
// where domain is the domain name of the server.
func (p packageConfig) goImportLink(domain string) string {
	return fmt.Sprintf("%v/%v %v", domain, p.name(), p.goMetaContent())
}

// goImportMeta creates the <meta/> HTML tag containing name and content attributes.
func (p packageConfig) goImportMeta(domain string) string {
	link := p.goImportLink(domain)
	return fmt.Sprintf(`<meta name="go-import" content="%s">`, link)
}

func (s vanityServer) find(path string) *packageConfig {
	p, ok := s.Packages[path]
	if !ok {
		return nil
	}
	return p
}

// ServeHTTP is an HTTP Handler for Go vanity domain.
func (s vanityServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	if r.Method != http.MethodGet {
		status := http.StatusMethodNotAllowed
		http.Error(w, http.StatusText(status), status)
		return
	}

	pack := s.find(r.URL.Path)
	if pack == nil {
		http.NotFound(w, r)
		return
	}
	if r.FormValue("go-get") != "1" {
		url := pack.goDocURL(s.Domain)
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
		return
	}
	fmt.Fprint(w, pack.goImportMeta(s.Domain))
}