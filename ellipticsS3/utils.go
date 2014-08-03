package ellipticsS3

import (
	"net/http"
	"strings"
)

func GetAuth(f func(context Context, w http.ResponseWriter, r *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username := strings.Split(
			strings.Split(r.Header.Get("Authorization"), ":")[0],
			" ",
		)[1]
		context := Context{
			Username: username,
		}
		f(context, w, r)
	}
}
