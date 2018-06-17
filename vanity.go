package vanity

import (
	"bytes"
	"html/template"
	"net/http"
	"strings"
)

type Data struct {
	ImportRoot string
	VCS        string
	VCSRoot    string
}

var tmpl = template.Must(template.New("main").Parse(`<!DOCTYPE html>
<html>
<head>
<meta http-equiv="Content-Type" content="text/html; charset=utf-8"/>
<meta name="go-import" content="{{.ImportRoot}} {{.VCS}} https://{{.VCSRoot}}">
</head>
</html>
`))

// Redirect is a HTTP middleware that redirects browsers to godoc.org or
// Go tool to VCS repository.
func Redirect(vcs, importPath, repoRoot string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Scheme == "http" {
			r.URL.Scheme = "https"
			http.Redirect(w, r, r.URL.String(), http.StatusMovedPermanently)
			return
		}
		if r.Method != http.MethodGet {
			status := http.StatusMethodNotAllowed
			http.Error(w, http.StatusText(status), status)
			return
		}

		if !strings.HasPrefix(strings.TrimSuffix(r.Host+r.URL.Path, "/") + "/", importPath+"/") {
			http.NotFound(w, r)
			return
		}
		if r.FormValue("go-get") != "1" {
			url := "https://godoc.org/" + r.Host + r.URL.Path
			http.Redirect(w, r, url, http.StatusTemporaryRedirect)
			return
		}

		d := &Data{
			ImportRoot: importPath,
			VCS:        vcs,
			VCSRoot:    repoRoot,
		}
		var buf bytes.Buffer
		err := tmpl.Execute(&buf, d)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Cache-Control", "public, max-age=300")
		w.Write(buf.Bytes())
	})
}
