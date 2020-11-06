package http

import (
	"net"
	"net/http"
	"testing"
	"time"
)

// This test will sleep for 50ms after the request is made asynchronously via RetryableClient
// as to test that the exponential backoff is working.
func TestClient_Delete_Backoff(t *testing.T) {
	// Collect errors for goroutines
	errs := make(chan error)

	// Mock server which always responds 200.
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Fatalf("bad method: %s", r.Method)
		}
		if r.RequestURI != "/foo/bar" {
			t.Fatalf("bad uri: %s", r.RequestURI)
		}
		w.WriteHeader(200)
	})

	doneCh := make(chan struct{})
	// Make the request asynchronously and deliberately fail the first request,
	// exponential backoff should pass on subsequent attempts of the same request
	go func() {
		defer close(doneCh)
		resp, err := NewRetryableClient().Delete("http://127.0.0.1:28934/foo/bar")
		if err != nil {
			errs <- err
		}
		resp.Body.Close()
	}()
	time.Sleep(50 * time.Millisecond)

	// Create a test server
	list, err := net.Listen("tcp", ":28934")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	go func() {
		defer list.Close()
		err = http.Serve(list, handler)
		if err != nil {
			errs <- err
		}
	}()

	<-doneCh
	// close errs once all children finish.
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatalf("err: %v", err)
		}
	}
}

// This test simulates a 404, to which the backoff should not apply as it should only care about networking problems.
// If the client is found to be retrying, the test will fail.
func TestClient_Get_Fail_Timeout(t *testing.T) {
	// Collect errors for goroutines
	errs := make(chan error)

	// Mock server which always responds 200.
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Fatalf("bad method: %s", r.Method)
		}
		if r.RequestURI != "/foo/bar" {
			t.Fatalf("bad uri: %s", r.RequestURI)
		}
		w.WriteHeader(404)
	})

	// Create a test server
	list, err := net.Listen("tcp", ":28935")
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	go func() {
		defer list.Close()
		err = http.Serve(list, handler)
		if err != nil {
			errs <- err
		}
	}()

	// Send the asynchronous request
	var resp *http.Response
	doneCh := make(chan struct{})
	go func() {
		defer close(doneCh)
		var err error
		resp, err = NewRetryableClient().Get("http://127.0.0.1:28935/foo/bar")
		if err != nil {
			errs <- err
		}
		resp.Body.Close()
	}()
	select {
	case <-doneCh:
		// close errs once all children finish.
		close(errs)
		for err := range errs {
			if err != nil {
				t.Fatalf("err: %v", err)
			}
		}
	case <-time.After(1000 * time.Millisecond):
		t.Fatalf("Client should fail instantly with a 404 response, not retrying!")
	}
}
