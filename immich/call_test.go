package immich

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

type testServer struct {
	// endpoint       string
	responseStatus int
	responseBody   string
}

func (ts *testServer) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	resp.WriteHeader(ts.responseStatus)
	resp.Write([]byte(ts.responseBody))
}

func TestCall(t *testing.T) {
	tt := []struct {
		name        string
		requestFn   requestFunction
		expectedErr bool
		server      testServer
	}{
		{
			name:        "happy path",
			requestFn:   get("/assets", setAcceptJSON()),
			expectedErr: false,
			server: testServer{
				responseStatus: http.StatusOK,
				responseBody:   `{"status": "All correct"}`,
			},
		},
		{
			name:        "bad url",
			requestFn:   get("/ass\nets", setAcceptJSON()),
			expectedErr: true,
			server: testServer{
				responseStatus: http.StatusOK,
				responseBody:   `{"status": "All correct"}`,
			},
		},
		{
			name:        "post / ok",
			requestFn:   post("/album", "application/json", setAcceptJSON(), setJSONBody(struct{ Name string }{Name: "test"})),
			expectedErr: false,
			server: testServer{
				responseStatus: http.StatusOK,
				responseBody:   `{"Name": "test"}`,
			},
		},
		{
			name:        "bad request / post",
			requestFn:   post("/album", "application/json", setAcceptJSON(), setJSONBody(struct{ Name string }{Name: "test"})),
			expectedErr: true,
			server: testServer{
				responseStatus: http.StatusBadRequest,
				responseBody:   `{"error": "Bad request", "statusCode": "400", "message": ["String1","String2"]}`,
			},
		},
	}

	for _, tst := range tt {
		t.Run(tst.name, func(t *testing.T) {
			server := httptest.NewServer(&tst.server)
			defer server.Close()
			ctx := context.Background()
			ic, err := NewImmichClient(server.URL, "1234", "tester", true)
			if err != nil {
				t.Fail()
				return
			}
			r := map[string]string{}
			err = ic.newServerCall(ctx, tst.name).do(tst.requestFn, responseJSON(&r))
			if tst.expectedErr && err == nil {
				t.Errorf("expected error, but no error")
			}
			if !tst.expectedErr && err != nil {
				t.Errorf("no error expected, but error: %s", err.Error())
			}
			if err != nil {
				t.Logf("error received: %s", err.Error())
			}
			t.Logf("response received: %#v", r)
		})
	}

}
