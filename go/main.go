package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/graphql-go/graphql"
)

type contextKey string

const headersKey contextKey = "headers"

func main() {
	httpPort := 9090

	schema, err := buildSchema()
	if err != nil {
		log.Fatalf("failed to build schema: %v", err)
	}

	http.HandleFunc("/", graphqlHandler(schema))

	// Kept as plain HTTP so container/platform probes keep working.
	http.HandleFunc("/healthz/", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		fmt.Fprintf(w, "{\"healthy\": true}")
	})

	fmt.Printf("listening on %v\n", httpPort)

	err = http.ListenAndServe(fmt.Sprintf(":%d", httpPort), logRequest(http.DefaultServeMux))
	if err != nil {
		log.Fatal(err)
	}
}

type graphqlRequest struct {
	Query         string                 `json:"query"`
	OperationName string                 `json:"operationName"`
	Variables     map[string]interface{} `json:"variables"`
}

func graphqlHandler(schema graphql.Schema) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		var gqlReq graphqlRequest

		switch req.Method {
		case http.MethodGet:
			gqlReq.Query = req.URL.Query().Get("query")
			gqlReq.OperationName = req.URL.Query().Get("operationName")
			if raw := req.URL.Query().Get("variables"); raw != "" {
				if err := json.Unmarshal([]byte(raw), &gqlReq.Variables); err != nil {
					writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid variables: %v", err))
					return
				}
			}
			if gqlReq.Query == "" {
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				fmt.Fprint(w, playgroundHTML)
				return
			}
		case http.MethodPost:
			body, err := io.ReadAll(req.Body)
			if err != nil {
				writeError(w, http.StatusInternalServerError, "failed to read request body")
				return
			}
			if err := json.Unmarshal(body, &gqlReq); err != nil {
				writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid GraphQL request: %v", err))
				return
			}
		default:
			writeError(w, http.StatusMethodNotAllowed, "Method not allowed. Use GET or POST.")
			return
		}

		result := graphql.Do(graphql.Params{
			Schema:         schema,
			RequestString:  gqlReq.Query,
			OperationName:  gqlReq.OperationName,
			VariableValues: gqlReq.Variables,
			Context:        context.WithValue(req.Context(), headersKey, req.Header),
		})

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(result); err != nil {
			log.Printf("failed to encode response: %v", err)
		}
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"errors": []map[string]string{{"message": message}},
	})
}

func buildSchema() (graphql.Schema, error) {
	bodyResultType := graphql.NewObject(graphql.ObjectConfig{
		Name: "BodyResult",
		Fields: graphql.Fields{
			"message":    &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"bodyLength": &graphql.Field{Type: graphql.NewNonNull(graphql.Int)},
		},
	})

	headersResultType := graphql.NewObject(graphql.ObjectConfig{
		Name: "HeadersResult",
		Fields: graphql.Fields{
			"message":     &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"headerCount": &graphql.Field{Type: graphql.NewNonNull(graphql.Int)},
			"headers":     &graphql.Field{Type: graphql.NewList(graphql.NewNonNull(graphql.String))},
		},
	})

	queryType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"active": &graphql.Field{
				Type: graphql.NewNonNull(graphql.Boolean),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return true, nil
				},
			},
			"healthy": &graphql.Field{
				Type: graphql.NewNonNull(graphql.Boolean),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return true, nil
				},
			},
			"hello": &graphql.Field{
				Type: graphql.NewNonNull(graphql.String),
				Args: graphql.FieldConfigArgument{
					"name": &graphql.ArgumentConfig{Type: graphql.String},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					name, _ := p.Args["name"].(string)
					return fmt.Sprintf("Hello %s", name), nil
				},
			},
			"printHeaders": &graphql.Field{
				Type: graphql.NewNonNull(headersResultType),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					headers, _ := p.Context.Value(headersKey).(http.Header)

					var headerParts []string
					for name, values := range headers {
						for _, value := range values {
							headerParts = append(headerParts, fmt.Sprintf("%s: %s", name, value))
						}
					}

					// Print all headers in a single line
					log.Printf("Request Headers: %s", strings.Join(headerParts, " | "))

					return map[string]interface{}{
						"message":     "Headers received and printed",
						"headerCount": len(headers),
						"headers":     headerParts,
					}, nil
				},
			},
			"five": &graphql.Field{
				Type: graphql.Int,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return nil, fmt.Errorf("deliberate failure: status 500")
				},
			},
			"four09": &graphql.Field{
				Type: graphql.Int,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return nil, fmt.Errorf("deliberate failure: status 409")
				},
			},
		},
	})

	mutationType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Mutation",
		Fields: graphql.Fields{
			"printBody": &graphql.Field{
				Type: graphql.NewNonNull(bodyResultType),
				Args: graphql.FieldConfigArgument{
					"body": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					body, _ := p.Args["body"].(string)

					// Print the body to console/logs
					log.Printf("Received body: %s", body)

					return map[string]interface{}{
						"message":    "Body received and printed",
						"bodyLength": len(body),
					}, nil
				},
			},
			"proxy": &graphql.Field{
				Type: graphql.NewNonNull(graphql.String),
				Args: graphql.FieldConfigArgument{
					"host": &graphql.ArgumentConfig{Type: graphql.String},
					"args": &graphql.ArgumentConfig{Type: graphql.String},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					host, _ := p.Args["host"].(string)
					args, _ := p.Args["args"].(string)

					if len(host) == 0 {
						host = "http://postman-echo.com"
					}
					if len(args) == 0 {
						args = "get?foo1=bar1&foo2=bar2"
					}

					resp, err := http.Get(fmt.Sprintf("%s/%s", strings.TrimRight(host, "/"), strings.TrimLeft(args, "/")))
					if err != nil {
						return nil, err
					}
					defer resp.Body.Close()

					body, err := io.ReadAll(resp.Body)
					if err != nil {
						return nil, err
					}
					return string(body), nil
				},
			},
		},
	})

	return graphql.NewSchema(graphql.SchemaConfig{
		Query:    queryType,
		Mutation: mutationType,
	})
}

func logRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}

const playgroundHTML = `<!doctype html>
<html>
  <head>
    <title>byoc-test GraphQL</title>
    <link rel="stylesheet" href="https://unpkg.com/graphiql@3/graphiql.min.css" />
  </head>
  <body style="margin:0;height:100vh;">
    <div id="graphiql" style="height:100vh;"></div>
    <script src="https://unpkg.com/react@18/umd/react.production.min.js"></script>
    <script src="https://unpkg.com/react-dom@18/umd/react-dom.production.min.js"></script>
    <script src="https://unpkg.com/graphiql@3/graphiql.min.js"></script>
    <script>
      const root = ReactDOM.createRoot(document.getElementById('graphiql'));
      root.render(
        React.createElement(GraphiQL, {
          fetcher: GraphiQL.createFetcher({ url: window.location.pathname }),
          defaultQuery: '{\n  active\n  hello(name: "choreo")\n}',
        })
      );
    </script>
  </body>
</html>
`

//trigger build 5
