package ellipticsS3

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	"github.com/noxiouz/elliptics-go/rift"
)

var (
	riftcli *rift.RiftClient
)

func bucketExists(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bucket := vars["bucket"]
	_, err := riftcli.ReadBucket(bucket)
	if err != nil {
		log.Println(err)
		http.Error(w, "Bucket doesn't exist", http.StatusNotFound)
		return
	}
	fmt.Fprintf(w, "OK")
}

func bucketCreate(context Context, w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented", http.StatusNotImplemented)
	return

	vars := mux.Vars(r)
	bucket := vars["bucket"]
	log.Printf("Create bucket. user: %s, bucket: %s", context.Username, bucket)
	if bucket == "" {
		http.Error(w, "bucket is undefined", http.StatusBadRequest)
	}

	// _, err := riftcli.CreateBucket(username, bucket)
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	return
	// }

	// fmt.Fprintf(w, "OK")
}

func objectGet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bucket := vars["bucket"]
	key := vars["key"]

	data, err := riftcli.GetObject(key, bucket, 0, 0)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	fmt.Fprintf(w, "%s", data)
}

func objectPut(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bucket := vars["bucket"]
	key := vars["key"]

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	defer r.Body.Close()

	// boto client check ETag header for proper MD5 summ
	h := md5.New()
	h.Write(data)
	etag := fmt.Sprintf("\"%x\"", h.Sum(nil))
	w.Header().Set("ETag", etag)

	_, err = riftcli.UploadObject(key, bucket, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	fmt.Fprintf(w, "OK")
}

func objectExists(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bucket := vars["bucket"]
	key := vars["key"]

	exists, err := riftcli.GetObject(key, bucket, 1, 0)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if len(exists) > 0 {
		fmt.Fprintf(w, "")
	} else {
		http.Error(w, "", http.StatusNotFound)
	}
}

func GetRouter(endpoint string) (h http.Handler, err error) {
	riftcli, err = rift.NewRiftClient(endpoint)
	if err != nil {
		return
	}
	//main router
	router := mux.NewRouter()
	router.StrictSlash(true)
	// buckets
	router.HandleFunc("/{bucket}/", bucketExists).Methods("HEAD")
	router.HandleFunc("/{bucket}/", GetAuth(bucketCreate)).Methods("PUT")
	// objects
	router.HandleFunc("/{bucket}/{key}", objectExists).Methods("HEAD")
	router.HandleFunc("/{bucket}/{key}", objectGet).Methods("GET")
	router.HandleFunc("/{bucket}/{key}", objectPut).Methods("PUT")
	// debug
	h = handlers.LoggingHandler(os.Stdout, router)
	return
}
