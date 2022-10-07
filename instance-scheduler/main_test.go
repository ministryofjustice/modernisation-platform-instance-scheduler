package main

import (
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestHandler(t *testing.T) {
	os.Setenv("INSTANCE_SCHEDULING_ACTION", "Test")

	t.Run("Successful Request", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
			fmt.Fprintf(w, "Testing")
		}))
		defer ts.Close()

		DefaultHTTPGetAddress = ts.URL

		_, err := handler(events.APIGatewayProxyRequest{})
		if err != nil {
			t.Fatal("Everything should be ok")
		}
	})
}
