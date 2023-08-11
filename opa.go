package traefikopa

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
)

type Config struct {
	URL string `json:"url,omitempty"`
}

func CreateConfig() *Config {
	return &Config{
		URL: "http://localhost:8181/v1/data/httpapi/authz", // Default OPA URL. overridden by middleware config, if provided
	}
}

type TraefikOPA struct {
	next   http.Handler
	name   string
	URL    string
	client *http.Client
}

func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	return &TraefikOPA{
		next:   next,
		name:   name,
		URL:    config.URL,
		client: &http.Client{},
	}, nil
}

func (o *TraefikOPA) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	input, err := o.constructInput(req)
	if err != nil {
		log.Println("Error constructing input:", err)
		http.Error(rw, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	allow, err := o.isAllowed(input)
	if err != nil {
		log.Println("Error checking OPA policy:", err)
		http.Error(rw, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if !allow {
		http.Error(rw, "Forbidden", http.StatusForbidden)
		return
	}

	o.next.ServeHTTP(rw, req)
}

func (o *TraefikOPA) constructInput(req *http.Request) (map[string]interface{}, error) {
	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	formParams := make(map[string]interface{})
	if req.Method == http.MethodPost {
		req.ParseForm()
		for key, values := range req.Form {
			formParams[key] = values[0]
		}
	}

	return map[string]interface{}{
		"http": map[string]interface{}{
			"method": req.Method,
			"scheme": req.URL.Scheme,
			"host":   req.Host,
			"path":   req.URL.Path,
			"headers": map[string]interface{}{
				"Authorization": req.Header.Get("Authorization"),
			},
			"body":         string(bodyBytes),
			"query_params": req.URL.Query(),
			"form_params":  formParams,
			"protocol":     req.Proto,
		},
	}, nil
}

func (o *TraefikOPA) isAllowed(input map[string]interface{}) (bool, error) {
	inputJSON, err := json.Marshal(input)
	if err != nil {
		return false, err
	}

	encodedInput := url.QueryEscape(string(inputJSON))
	opaURLWithParam := o.URL + "?input=" + encodedInput

	resp, err := o.client.Get(opaURLWithParam)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	var opaResponse struct {
		Result struct {
			Allow bool `json:"allow"`
		} `json:"result"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&opaResponse); err != nil {
		return false, err
	}

	return opaResponse.Result.Allow, nil
}
