package runner

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/nais/tester/lua/spec"
	lua "github.com/yuin/gopher-lua"
)

type GQL struct {
	server http.Handler

	results map[string]any
}

var _ spec.Runner = (*GQL)(nil)

func NewGQLRunner(server http.Handler) *GQL {
	return &GQL{server: server}
}

func (g *GQL) Name() string {
	return "gql"
}

func (g *GQL) Functions() []*spec.Function {
	return []*spec.Function{
		{
			Name: "query",
			Args: []spec.Argument{
				{
					Name: "query",
					Type: []spec.ArgumentType{spec.ArgumentTypeString},
					Doc:  "The query to run",
				},
				{
					Name: "headers?",
					Type: []spec.ArgumentType{spec.ArgumentTypeTable},
					Doc:  "The headers to add to the HTTP request",
				},
			},
			Doc:  "Query comment",
			Func: g.query,
		},
		StdCheckDefinition(g.check),
	}
}

func (g *GQL) query(L *lua.LState) int {
	query := L.CheckString(1)
	headers := L.OptTable(2, L.NewTable())

	rec := httptest.NewRecorder()

	b, err := json.Marshal(
		struct {
			OperationName *string        `json:"operationName"`
			Variables     map[string]any `json:"variables"`
			Query         string         `json:"query"`
		}{
			nil,
			map[string]any{},
			query,
		},
	)
	if err != nil {
		panic(fmt.Sprintf("gql.Run: unable to marshal request: %v", err))
	}

	req, err := http.NewRequestWithContext(L.Context(), "POST", "/", bytes.NewReader(b))
	if err != nil {
		panic(fmt.Sprintf("gql.Run: unable to create request: %v", err))
	}

	req.Header.Add("Content-Type", "application/json")

	headers.ForEach(func(k, v lua.LValue) {
		req.Header.Add(k.String(), v.String())
	})

	g.server.ServeHTTP(rec, req)

	g.results = map[string]any{}
	if err := json.Unmarshal(rec.Body.Bytes(), &g.results); err != nil {
		panic(fmt.Sprintf("gql.Run: unable to unmarshal response: %v", err))
	}

	return 0
}

func (g *GQL) check(L *lua.LState) int {
	tbl := L.CheckTable(1)
	StdCheck(L, tbl, g.results)
	return 0
}

// func (g *GQL) Run(ctx context.Context, logf func(format string, args ...any), body []byte, state map[string]any) error {
// 	f, err := parser.Parse(body, state)
// 	if err != nil {
// 		return fmt.Errorf("gql.Parse: %w", err)
// 	}

// 	return f.Execute(state, func() (any, error) {
// 	})
// }
