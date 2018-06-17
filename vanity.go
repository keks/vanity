package vanity

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/pkg/errors"
)

func Redirect(vcs, importRoot, repoRoot string) http.Handler {
	return &Repo{
		ImportPrefix: importRoot,
		VCS: vcs,
		RepoRoot: repoRoot,
	}
}

type Repo struct {
	ImportPrefix string
	VCS        string
	RepoRoot   string
}

var tmpl = template.Must(template.New("main").Parse(`<!DOCTYPE html>
<html>
<head>
<meta http-equiv="Content-Type" content="text/html; charset=utf-8"/>
<meta name="go-import" content="{{.ImportPrefix}} {{.VCS}} {{.RepoRoot}}">
</head>
</html>
`))

func (repo *Repo) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

	if !strings.HasPrefix(strings.TrimSuffix(r.Host+r.URL.Path, "/")+"/", repo.ImportPrefix+"/") {
		http.NotFound(w, r)
		return
	}
	if r.FormValue("go-get") != "1" {
		url := "https://godoc.org/" + r.Host + r.URL.Path
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
		return
	}

	var buf bytes.Buffer
	err := tmpl.Execute(&buf, repo)
	if err != nil {
		writeError(w, err)
		return
	}
	w.Header().Set("Cache-Control", "public, max-age=300")
	w.Write(buf.Bytes())
}

func HandleImports(imports []*Repo) http.Handler {
	defer fmt.Println("HandleImports initialized")
	importMap := make(map[string]*Repo)

	for _, imprt := range imports {
		importMap[imprt.ImportPrefix] = imprt
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := r.Host + r.URL.Path
		fmt.Println("incoming request:", p)

		var lastp string

		for p != lastp {
			fmt.Println(p)
			if imprt, ok := importMap[p]; ok {
				fmt.Printf("found %q for %q", p, r.URL.Path)
				imprt.ServeHTTP(w, r)
				return
			}
			lastp = p
			p, _ = path.Split(p)
			p = strings.TrimSuffix(p, "/")
		}

		fmt.Println(importMap)
		http.Error(w, "package not found", http.StatusNotFound)
	})
}

func HandleLoadFile(path string) (http.Handler, error) {
	defer fmt.Println("HandleLoadFile initialized")
	f, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrap(err, "error opening file")
	}

	var imports []*Repo

	rr := csv.NewReader(f)
	for {
		rec, err := rr.Read()
		if err != nil {
			if err == io.EOF {
				break
			}

			return nil, errors.Wrap(err, "error reading file")
		}

		if len(rec) < 3 {
			return nil, errors.New("invalid record length, need at least three")
		}

		imports = append(imports, &Repo{
			ImportPrefix:     rec[0],
			VCS:      rec[1],
			RepoRoot: rec[2],
		})
	}

	fmt.Printf("loaded imports: %#v\n", imports)

	return HandleImports(imports), nil
}

func HandleHotReloadFile(path string) http.Handler {
	var (
		lastLoad time.Time
		h        http.Handler
	)

	defer fmt.Println("HandleHotReloadFile initialized")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fi, err := os.Stat(path)
		if err != nil {
			writeError(w, err)
			return
		}

		if fi.ModTime().After(lastLoad) {
			var err error
			h, err = HandleLoadFile(path)
			if err != nil {
				writeError(w, err)
				return
			}
		}

		h.ServeHTTP(w, r)
	})
}

func writeError(w http.ResponseWriter, err error) {
	http.Error(w, "error performing request: "+err.Error(), http.StatusInternalServerError)
}
