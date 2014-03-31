package ellipticsS3

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

var (
	rift S3Backend
)

func BucketExists(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Not implemented")
}

func ObjectGet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bucket := vars["bucket"]
	key := vars["key"]

	data, err := rift.GetObject(key, bucket)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	fmt.Fprintf(w, "%s", data)
}

func ObjectPut(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bucket := vars["bucket"]
	key := vars["key"]

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	defer r.Body.Close()

	err = rift.UploadObject(bucket, key, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	fmt.Fprintf(w, "OK")
}

func GetRouter(endpoint string) (h http.Handler, err error) {
	rift, err = NewRiftbackend(endpoint)
	if err != nil {
		return
	}
	//main router
	router := mux.NewRouter()
	router.StrictSlash(true)
	// buckets
	router.HandleFunc("/{bucket}/", BucketExists).Methods("HEAD")
	// objects
	// router.HandleFunc("/{bucket}/{key}", ObjectExists).Methods("HEAD")
	router.HandleFunc("/{bucket}/{key}", ObjectGet).Methods("GET")
	router.HandleFunc("/{bucket}/{key}", ObjectPut).Methods("PUT")
	h = handlers.LoggingHandler(os.Stdout, router)
	return
}
