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
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"cloud.google.com/go/logging"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestHandler(t *testing.T) {
	ctx := context.Background()
	rr := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "http://example.com", strings.NewReader(`{
		"requestId": "124ab1c",
		"caller": "//bigquery.googleapis.com/projects/myproject/jobs/myproject:US.bquxjob_5b4c112c_17961fafeaf",
		"sessionUser": "test-user@test-company.com",
		"userDefinedContext": {
		 "key1": "value1",
		 "key2": "v2"
		},
		"calls": [
		 ["/wYAAFMyc1R3TwEFAAB4buQoYQ\u003d\u003d"],
		 ["/wYAAFMyc1R3TwEFAABh9epjdg\u003d\u003d"]
		]
	 }`))
	client, err := logging.NewClient(ctx, "projects/testing",
		option.WithoutAuthentication(),
		option.WithGRPCDialOption(
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		),
	)
	if err != nil {
		t.Fatalf("unable to initialize logging client: %v", err)
	}
	app := &App{
		log: client.Logger("test-log", logging.RedirectAsJSON(os.Stderr)),
	}

	app.Handler(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("Code = %d, want %d", rr.Code, http.StatusOK)
	}
	wantBody := "Hello World!\n"
	if got := rr.Body.String(); got != wantBody {
		t.Errorf("Body = %q, want %q", got, wantBody)
	}
}
