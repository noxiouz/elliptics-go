package ellipticsS3

import (
	"net/http"
	"strings"
)

func GetAuth(f func(username string, w http.ResponseWriter, r *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username := strings.Split(r.Header.Get("Authorization"), ":")[0]
		f(username, w, r)
	}
}
