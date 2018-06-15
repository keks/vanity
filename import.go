package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/keks/vanity"
)

type Import struct {
	Path     string // import path without domain
	VCS      string // usuall "git"
	RepoRoot string

	h http.Handler
}

func (i *Import) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if i.h == nil {
		i.h = vanity.Redirect(i.VCS, i.Path, i.RepoRoot)
		fmt.Println("Import Handler initialized for", i.Path)
	}

	i.h.ServeHTTP(w, r)
}

func HandleImports(imports []*Import) http.Handler {
	defer fmt.Println("HandleImports initialized")
	importMap := make(map[string]*Import)

	for _, imprt := range imports {
		importMap[imprt.Path] = imprt
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
