// Copyright 2022 Linka Cloud  All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package react

import (
	"io/fs"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
)

const (
	EndpointEnv = "REACT_ENDPOINT"
)

func NewHandler(dir fs.FS, subpath string) (http.Handler, error) {
	if e := os.Getenv(EndpointEnv); e != "" {
		return newProxy(e)
	}
	return newStatic(dir, subpath)
}

func newStatic(dir fs.FS, subpath string) (http.Handler, error) {
	s, err := fs.Sub(dir, subpath)
	if err != nil {
		return nil, err
	}
	fsrv := http.FileServer(http.FS(s))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := strings.TrimPrefix(r.URL.Path, "/")
		if _, err := fs.Stat(s, p); err != nil {
			if _, err := fs.Stat(s, p+".html"); err == nil {
				r.URL.Path += ".html"
			} else {
				r.URL.Path = "/"
			}
		}
		fsrv.ServeHTTP(w, r)
	}), nil
}

func newProxy(endpoint string) (http.Handler, error) {
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	p := httputil.NewSingleHostReverseProxy(u)
	return p, nil
}

func DevEnv() bool {
	return os.Getenv(EndpointEnv) != ""
}
