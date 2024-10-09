package runner

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/nais/tester/lua/spec"
	lua "github.com/yuin/gopher-lua"
)

type REST struct {
	server   http.Handler
	response *httptest.ResponseRecorder
}

var _ spec.Runner = (*REST)(nil)

func NewRestRunner(server http.Handler) *REST {
	return &REST{server: server}
}

func (r *REST) Name() string {
	return "rest"
}

func (s *REST) Functions() []*spec.Function {
	return []*spec.Function{
		{
			Name: "send",
			Args: []spec.Argument{
				{
					Name: "method",
					Type: []spec.ArgumentType{spec.StringEnum{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS", "HEAD"}},
					Doc:  "HTTP method",
				},
				{
					Name: "path",
					Type: []spec.ArgumentType{spec.ArgumentTypeString},
					Doc:  "The path to query",
				},
				{
					Name: "body?",
					Type: []spec.ArgumentType{spec.ArgumentTypeString, spec.ArgumentTypeTable},
					Doc:  "The body to send",
				},
			},
			Doc:  "Send http request",
			Func: s.send,
		},
		{
			Name: "check",
			Args: []spec.Argument{
				{
					Name: "status_code",
					Type: []spec.ArgumentType{spec.ArgumentTypeNumber},
					Doc:  "Expected status code",
				},
				{
					Name: "resp",
					Type: []spec.ArgumentType{spec.ArgumentTypeTable},
					Doc:  "Expected response",
				},
			},
			Doc:  "Check the response done by send",
			Func: s.check,
		},
	}
}

func (r *REST) send(L *lua.LState) int {
	if r.response != nil {
		L.RaiseError("send already called")
		return 0
	}

	ctx := L.Context()
	method := L.CheckString(1)
	path := L.CheckString(2)
	var body io.Reader
	if L.GetTop() > 2 {
		switch L.Get(3).(type) {
		case lua.LString:
			body = strings.NewReader(L.CheckString(3))
		case *lua.LTable:
			tbl := L.CheckTable(3)
			b, err := json.Marshal(tbl)
			if err != nil {
				L.RaiseError("unable to marshal table: %v", err)
			}
			body = bytes.NewReader(b)
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, path, body)
	if err != nil {
		panic(fmt.Errorf("rest.Run: unable to create request: %w", err))
	}

	r.response = httptest.NewRecorder()
	r.server.ServeHTTP(r.response, req)

	return 0
}

func (r *REST) check(L *lua.LState) int {
	code := L.CheckInt(1)
	tbl := L.CheckTable(2)

	if r.response == nil {
		L.RaiseError("send not called")
		return 0
	}

	if r.response.Code != code {
		L.RaiseError("expected response code %d, got %d\n%v", code, r.response.Code, r.response.Body.String())
		return 0
	}

	var res map[string]interface{}
	if err := json.Unmarshal(r.response.Body.Bytes(), &res); err != nil {
		L.RaiseError("unable to unmarshal response: %v", err)
		return 0
	}

	StdCheck(L, tbl, res)
	return 0
}

// func (s *REST) Run(ctx context.Context, logf func(format string, args ...any), body []byte, state map[string]any) error {
// 	f, err := parser.Parse(body, state)
// 	if err != nil {
// 		return fmt.Errorf("gql.Parse: %w", err)
// 	}

// 	return f.Execute(state, func() (any, error) {
// 		if rowTypeRegexp.MatchString(f.Query) {
// 			return s.queryRow(ctx, f)
// 		}
// 		return s.query(ctx, f)
// 	})
// }

// func (r *REST) Run(ctx context.Context, logf func(format string, args ...any), body []byte, state map[string]any) error {
// 	f, err := parser.Parse(body, state)
// 	if err != nil {
// 		return fmt.Errorf("rest.Parse: %w", err)
// 	}

// 	expectedResponseCode := f.Opts["responseCode"]

// 	delete(f.Opts, "responseCode")

// 	return f.Execute(state, func() (any, error) {
// 		rec := httptest.NewRecorder()

// 		parts := strings.SplitN(f.Query, "\n", 2)
// 		methodPath := strings.SplitN(strings.TrimSpace(parts[0]), " ", 2)

// 		logf("running %v request to %v", methodPath[0], methodPath[1])

// 		var body io.Reader
// 		if len(parts) > 1 {
// 			body = strings.NewReader(parts[1])
// 		}

// 		req, err := http.NewRequestWithContext(ctx, methodPath[0], methodPath[1], body)
// 		if err != nil {
// 			return nil, fmt.Errorf("rest.Run: unable to create request: %w", err)
// 		}

// 		r.server.ServeHTTP(rec, req)

// 		if expectedResponseCode != "" {
// 			if strconv.Itoa(rec.Code) != expectedResponseCode {
// 				return nil, fmt.Errorf("expected response code %q, got %d\n%v", expectedResponseCode, rec.Code, rec.Body.String())
// 			}
// 		}

// 		res := map[string]any{}
// 		if err := yaml.Unmarshal(rec.Body.Bytes(), &res); err != nil {
// 			return nil, fmt.Errorf("rest.Run: unable to unmarshal response: %w", err)
// 		}

// 		return res, nil
// 	})
// }
