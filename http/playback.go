package fbhttp

import (
	"net/http"
	"time"
)

// ponytail: 30 min is the ceiling a leaked playback token stays usable; the
// token is single-file and never renewed, so the blast radius of a copy in a
// proxy log or browser history stays small.
const playbackTokenExpiration = 30 * time.Minute

// playbackTokenHandler issues a short-lived token scoped to a single raw
// file, for external video players that can only authenticate via ?auth= in
// the URL.
var playbackTokenHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	if !d.user.Perm.Download {
		return http.StatusForbidden, nil
	}

	rawPath := slashClean(r.URL.Query().Get("path"))
	if rawPath == "" || rawPath == "/" {
		return http.StatusBadRequest, nil
	}

	return printToken(w, r, d, d.user, playbackTokenExpiration, rawPath)
})
