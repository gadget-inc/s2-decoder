// Copyright 2021 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	"cloud.google.com/go/logging"
	"github.com/klauspost/compress/s2"
)

type query struct {
	RequestID   string     `json:"request_id"`
	Caller      string     `json:"caller"`
	SessionUser string     `json:"sessionUser"`
	Calls       [][]string `json:"calls"`
}

type response struct {
	Replies []string `json:"replies"`
}

func decodeStream(src []byte) (string, error) {
	dec := s2.NewReader(bytes.NewReader(src))
	buf := new(strings.Builder)
	_, err := io.Copy(buf, dec)

	return buf.String(), err
}

func (a *App) Handler(w http.ResponseWriter, r *http.Request) {
	var query query

	err := json.NewDecoder(r.Body).Decode(&query)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	a.log.Log(logging.Entry{
		Severity: logging.Info,
		HTTPRequest: &logging.HTTPRequest{
			Request: r,
		},
		Labels:  map[string]string{"caller": query.Caller, "length": strconv.Itoa(len(query.Calls))},
		Payload: "processing request",
	})

	results := make([]string, len(query.Calls))

	for i, s := range query.Calls {
		compressed, err := base64.StdEncoding.DecodeString(s[0])

		decoded, err := decodeStream(compressed)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		results[i] = decoded
	}

	response := response{Replies: results}
	jsonResponse, jsonError := json.Marshal(response)

	if jsonError != nil {
		a.log.Log(logging.Entry{
			Severity: logging.Info,
			HTTPRequest: &logging.HTTPRequest{
				Request: r,
			},
			Payload: "unable to encode json",
		})
		http.Error(w, err.Error(), http.StatusBadRequest)
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(jsonResponse)
	}

}
