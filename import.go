package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/pkg/errors"
	"kkn.fi/vanity"
)

type Import struct {
	Path     string // full import path
	VCS      string // usuall "git"
	RepoRoot string

	h http.Handler
}

func (i *Import) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if i.h == nil {
		i.h = vanity.Redirect(i.VCS, i.Path, i.RepoRoot)
	}

	i.h.ServeHTTP(w, r)
}

func HandleImports(imports []*Import) http.Handler {
	importMap := make(map[string]*Import)

	for _, imprt := range imports {
		importMap[imprt.Path] = imprt
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := r.URL.Host + "/" + r.URL.Path

		for len(p) > 0 {
			if imprt, ok := importMap[p]; ok {
				fmt.Printf("found %q for %q", p, r.URL.Path)
				imprt.ServeHTTP(w, r)
				return
			}
			p, _ = path.Split(p)
		}

		http.Error(w, "package not found", http.StatusNotFound)
	})
}

func HandleLoadFile(path string) (http.Handler, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrap(err, "error opening file")
	}

	var imports []*Import

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

		imports = append(imports, &Import{
			Path:     rec[0],
			VCS:      rec[1],
			RepoRoot: rec[2],
		})
	}

	return HandleImports(imports), nil
}

func HandleHotReloadFile(path string) http.Handler {
	var (
		lastLoad time.Time
		h        http.Handler
	)

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
