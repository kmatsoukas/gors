package gors

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// The path is sent as written: an encoded segment (a file id that is itself a path, for
// instance) must reach the server as %2F and not as a re-escaped %252F.
func TestSendKeepsThePathAsTheCallerWroteIt(t *testing.T) {
	tests := []struct {
		name     string
		basePath string
		path     string
		expected string
	}{
		{"encoded segment", "/api", "/files/movies%2Fmovie.mkv/download", "/api/files/movies%2Fmovie.mkv/download"},
		{"unencoded reserved characters", "/api", "/files/a b#c", "/api/files/a%20b%23c"},
		{"stray percent", "/api", "/files/100%", "/api/files/100%25"},
		{"trailing slash", "/api", "/files/abc/", "/api/files/abc/"},
		{"no base path", "", "/files/abc", "/files/abc"},
		{"relative path", "/api", "files/abc", "/api/files/abc"},
		{"dirty path", "/api/", "//files//abc", "/api/files/abc"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var got string

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				got = r.URL.EscapedPath()
			}))
			defer server.Close()

			client := NewClient(server.URL + test.basePath)

			res, err := client.NewRequest(GET, test.path).Send()
			if err != nil {
				t.Fatalf("sending the request: %v", err)
			}
			res.Body.Close()

			if got != test.expected {
				t.Errorf("unexpected path:\n got: %s\nwant: %s", got, test.expected)
			}
		})
	}
}

func TestSendAddsTheQueryParameters(t *testing.T) {
	var got string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.URL.RawQuery
	}))
	defer server.Close()

	client := NewClient(server.URL)

	req := client.NewRequest(GET, "/files")
	req.SetQuery("parent", "movies/2024")

	res, err := req.Send()
	if err != nil {
		t.Fatalf("sending the request: %v", err)
	}
	res.Body.Close()

	if expected := "parent=movies%2F2024"; got != expected {
		t.Errorf("unexpected query:\n got: %s\nwant: %s", got, expected)
	}
}
