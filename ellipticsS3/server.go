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
	riftcli     *rift.RiftClient
	globalConfg Config
)

func getAllBuckets(context Context, w http.ResponseWriter, r *http.Request) {
	listing, err := riftcli.ListBucketDirectory(context.Username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Println("Get all %s buckets: %v", context.Username, listing)
	fmt.Fprint(w, listing)
}

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
	vars := mux.Vars(r)
	bucket := vars["bucket"]
	log.Printf("Create bucket. user: %s, bucket: %s, groups: %v", context.Username, bucket, globalConfg.DataGroups)
	if bucket == "" {
		http.Error(w, "bucket is undefined", http.StatusBadRequest)
	}

	bucketOpt := rift.BucketOptions{
		Groups:    globalConfg.DataGroups,
		ACL:       make([]rift.ACLStruct, 0),
		Flags:     0,
		MaxSize:   0,
		MaxKeyNum: 0,
	}

	info, err := riftcli.CreateBucket(bucket, context.Username, bucketOpt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Bucket %s has been created %v", bucket, info)
	fmt.Fprintf(w, "OK")
}

func bucketList(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bucket := vars["bucket"]
	log.Printf("List directory %s", bucket)
	listing, err := riftcli.ListBucket(bucket)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	log.Printf("List of directory %s: %s", bucket, listing)
	fmt.Println(w, "OK")
}

func objectGet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bucket := vars["bucket"]
	key := vars["key"]

	data, err := riftcli.GetObject(bucket, key, 0, 0)
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

	// boto client checks ETag header to verify MD5 summ
	h := md5.New()
	h.Write(data)
	etag := fmt.Sprintf("\"%x\"", h.Sum(nil))
	w.Header().Set("ETag", etag)

	_, err = riftcli.UploadObject(bucket, key, data)
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

	exists, err := riftcli.GetObject(bucket, key, 1, 0)
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

func GetRouter(config Config) (h http.Handler, err error) {
	globalConfg = config
	riftcli, err = rift.NewRiftClient(globalConfg.Endpoint)
	if err != nil {
		return
	}
	//main router
	router := mux.NewRouter()
	router.StrictSlash(true)
	// buckets
	router.HandleFunc("/{bucket}/", bucketExists).Methods("HEAD")
	router.HandleFunc("/{bucket}/", bucketList).Methods("GET")
	router.HandleFunc("/{bucket}/", GetAuth(bucketCreate)).Methods("PUT")
	router.HandleFunc("/", GetAuth(getAllBuckets)).Methods("GET")
	// objects
	router.HandleFunc("/{bucket}/{key}", objectExists).Methods("HEAD")
	router.HandleFunc("/{bucket}/{key}", objectGet).Methods("GET")
	router.HandleFunc("/{bucket}/{key}", objectPut).Methods("PUT")
	// debug
	h = handlers.LoggingHandler(os.Stdout, router)
	return
}
