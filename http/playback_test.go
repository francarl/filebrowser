package fbhttp

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/users"
)

func TestPlaybackTokenScopesRawFile(t *testing.T) {
	key := []byte("test-signing-key")
	perm := users.Permissions{Download: true}

	userScope := t.TempDir()
	for name, content := range map[string]string{"movie.mp4": "movie-bytes", "other.mp4": "other-bytes"} {
		if err := os.WriteFile(filepath.Join(userScope, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	st := scopedUserStorage(t, userScope, perm, key)
	session := signToken(t, perm, key)

	req, _ := http.NewRequest(http.MethodGet, "/api/playback-token?path=/movie.mp4", http.NoBody)
	req.Header.Set("X-Auth", session)
	issued := httptest.NewRecorder()
	handle(playbackTokenHandler, "", st, &settings.Server{}).ServeHTTP(issued, req)
	if issued.Code != http.StatusOK {
		t.Fatalf("issue: got %d, want 200", issued.Code)
	}
	token := strings.TrimSpace(issued.Body.String())

	var tk authToken
	if _, err := jwt.NewParser().ParseWithClaims(token, &tk, func(*jwt.Token) (interface{}, error) {
		return key, nil
	}); err != nil {
		t.Fatalf("parse playback token: %v", err)
	}
	if tk.RawPath != "/movie.mp4" {
		t.Fatalf("RawPath = %q, want /movie.mp4", tk.RawPath)
	}
	if tk.ExpiresAt == nil || time.Until(tk.ExpiresAt.Time) > playbackTokenExpiration {
		t.Fatalf("playback token not bounded by playbackTokenExpiration")
	}

	raw := func(path, authToken string) *httptest.ResponseRecorder {
		req, _ := http.NewRequest(http.MethodGet, "/api/raw"+path, http.NoBody)
		req.URL.RawQuery = "auth=" + authToken
		rec := httptest.NewRecorder()
		handle(rawHandler, "/api/raw", st, &settings.Server{}).ServeHTTP(rec, req)
		return rec
	}

	if rec := raw("/movie.mp4", token); rec.Code != http.StatusOK || rec.Body.String() != "movie-bytes" {
		t.Errorf("scoped raw = %d %q; want 200 movie-bytes", rec.Code, rec.Body.String())
	}
	if rec := raw("/other.mp4", token); rec.Code != http.StatusForbidden {
		t.Errorf("out-of-scope raw = %d; want 403", rec.Code)
	}

	req, _ = http.NewRequest(http.MethodGet, "/api/renew", http.NoBody)
	req.Header.Set("X-Auth", token)
	rec := httptest.NewRecorder()
	handle(renewHandler(time.Hour), "", st, &settings.Server{}).ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Errorf("playback token on /api/renew = %d; want 403", rec.Code)
	}
}

func TestPlaybackTokenRequiresDownload(t *testing.T) {
	key := []byte("test-signing-key")
	st := scopedUserStorage(t, t.TempDir(), users.Permissions{}, key)

	req, _ := http.NewRequest(http.MethodGet, "/api/playback-token?path=/movie.mp4", http.NoBody)
	req.Header.Set("X-Auth", signToken(t, users.Permissions{}, key))
	rec := httptest.NewRecorder()
	handle(playbackTokenHandler, "", st, &settings.Server{}).ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Errorf("issue without download perm = %d; want 403", rec.Code)
	}
}

func TestPlaybackTokenRequiresPath(t *testing.T) {
	key := []byte("test-signing-key")
	perm := users.Permissions{Download: true}
	st := scopedUserStorage(t, t.TempDir(), perm, key)

	req, _ := http.NewRequest(http.MethodGet, "/api/playback-token", http.NoBody)
	req.Header.Set("X-Auth", signToken(t, perm, key))
	rec := httptest.NewRecorder()
	handle(playbackTokenHandler, "", st, &settings.Server{}).ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("issue without path = %d; want 400", rec.Code)
	}
}
