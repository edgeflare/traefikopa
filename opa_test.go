package traefikopa

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTraefikOPA(t *testing.T) {
	tests := []struct {
		name        string
		opaResponse func(w http.ResponseWriter, r *http.Request)
		requests    []struct {
			method, path, body, authHeader string
			expectedCode                   int
		}
	}{
		{
			name:        "Test Policy Decision from Body",
			opaResponse: mockOPAServerWithRequestBody,
			requests: []struct {
				method, path, body, authHeader string
				expectedCode                   int
			}{
				{"POST", "/test", `{"profileID": "12345", "username": "testuser", "role": "user"}`, "12345", http.StatusOK},
				{"POST", "/test", `{"profileID": "67890", "username": "testuser", "role": "user"}`, "not_67890", http.StatusForbidden},
			},
		},
		{
			name:        "Test Method and Path",
			opaResponse: mockOPAServerWithMethodPath,
			requests: []struct {
				method, path, body, authHeader string
				expectedCode                   int
			}{
				{"GET", "/test", "", "", http.StatusOK},
				{"POST", "/test", "", "", http.StatusForbidden},
				{"GET", "/forbidden", "", "", http.StatusForbidden},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			opaServer, tearDown := createMockOPAServer(test.opaResponse)
			defer tearDown()

			middleware := setupMiddleware(opaServer.URL)

			for _, reqTest := range test.requests {
				req := httptest.NewRequest(reqTest.method, reqTest.path, strings.NewReader(reqTest.body))
				req.Header.Set("Authorization", reqTest.authHeader)
				rr := httptest.NewRecorder()

				middleware.ServeHTTP(rr, req)

				if rr.Code != reqTest.expectedCode {
					t.Errorf("Expected status %v but received %v", reqTest.expectedCode, rr.Code)
				}
			}
		})
	}
}

func setupMiddleware(opaURL string) http.Handler {
	config := CreateConfig()
	config.URL = opaURL

	stubNextHandler := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)
	})

	middleware, err := New(context.TODO(), stubNextHandler, config, "test-middleware")
	if err != nil {
		panic("Failed to create middleware: " + err.Error())
	}

	return middleware
}

func createMockOPAServer(handlerFunc func(w http.ResponseWriter, r *http.Request)) (*httptest.Server, func()) {
	server := httptest.NewServer(http.HandlerFunc(handlerFunc))
	return server, server.Close
}

func mockOPAServerWithRequestBody(w http.ResponseWriter, r *http.Request) {

	inputQuery := r.URL.Query().Get("input")
	var inputData map[string]interface{}
	err := json.Unmarshal([]byte(inputQuery), &inputData)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	authzHeader, _ := inputData["http"].(map[string]interface{})["headers"].(map[string]interface{})["Authorization"].(string)
	body, _ := inputData["http"].(map[string]interface{})["body"].(string)
	var bodyData map[string]interface{}
	json.Unmarshal([]byte(body), &bodyData)
	profileID, _ := bodyData["profileID"].(string)

	if authzHeader == profileID {
		w.Write([]byte(`{"result": {"allow": true}}`))
	} else {
		// Default deny
		w.Write([]byte(`{"result": {"allow": false}}`))
	}
}

func mockOPAServerWithMethodPath(w http.ResponseWriter, r *http.Request) {
	inputQuery := r.URL.Query().Get("input")
	var inputData map[string]interface{}
	err := json.Unmarshal([]byte(inputQuery), &inputData)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	method, _ := inputData["http"].(map[string]interface{})["method"].(string)
	path, _ := inputData["http"].(map[string]interface{})["path"].(string)

	if method == "GET" && path == "/test" {
		w.Write([]byte(`{"result": {"allow": true}}`))
	} else if path == "/forbidden" {
		w.Write([]byte(`{"result": {"allow": false}}`))
	} else {
		w.Write([]byte(`{"result": {"allow": false}}`))
	}
}
