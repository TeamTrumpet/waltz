package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/TeamTrumpet/waltz/waltz"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/codegangsta/negroni"
	"github.com/facebookgo/grace/gracehttp"
)

const imageExpiry time.Duration = 1 * time.Hour * 24 * 30

type resizeOpts struct {
	region, bucket, prefix string
	path                   string

	client *s3.S3
}

func newResizeHandler(ro resizeOpts) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Starting resizeHandler")

		imageKey := r.URL.Path[len(ro.path):]

		// fetch the query param
		resizeQuery := r.URL.Query().Get("resize")

		key := fmt.Sprintf("%s/%s", ro.prefix, imageKey)

		if resizeQuery == "" {
			log.Println("Completed resizeHandler: Redirected, no resize parameter")
			urlStr := fmt.Sprintf("https://s3-%s.amazonaws.com/%s/%s", ro.region, ro.bucket, key)
			http.Redirect(w, r, urlStr, http.StatusTemporaryRedirect)
			return
		}

		resizeX, resizeY, err := waltz.ParseResize(resizeQuery)
		if err != nil {
			log.Println("Completed resizeHandler: ERROR: Resize parameters invalid:", err.Error())
			http.Error(w, "", http.StatusBadRequest)
			return
		}

		params := &s3.GetObjectInput{
			Bucket: aws.String(ro.bucket),
			Key:    aws.String(key),
		}

		log.Println("Getting", *params.Key, "from bucket", *params.Bucket)

		ifNoneMatch := r.Header.Get("If-None-Match")
		if ifNoneMatch != "" {
			params.IfNoneMatch = aws.String(ifNoneMatch)
		}

		resp, err := ro.client.GetObject(params)
		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok {

				if reqErr, ok := err.(awserr.RequestFailure); ok {

					// if this is a not modified error, then cut quick
					if reqErr.StatusCode() == http.StatusNotModified {
						log.Println("Completed resizeHandler: Object not modified")
						w.WriteHeader(http.StatusNotModified)
						return
					}

					if reqErr.StatusCode() == http.StatusNotFound {
						log.Println("Completed resizeHandler: Key not found")
						w.WriteHeader(http.StatusNotFound)
						return
					}

					// A service error occurred
					log.Println("Completed resizeHandler: ERROR: AWS Service error:", reqErr.Code(), reqErr.Message(), reqErr.StatusCode(), reqErr.RequestID())
					http.Error(w, "", http.StatusInternalServerError)
					return

				}

				// Generic AWS error with Code, Message, and original error (if any)
				log.Println("Completed resizeHandler: ERROR: Generic AWS error:", awsErr.Code(), awsErr.Message(), awsErr.OrigErr())
				http.Error(w, "", http.StatusInternalServerError)
				return

			}

			// This case should never be hit, the SDK should always return an
			// error which satisfies the awserr.Error interface.
			log.Println("Completed resizeHandler: ERROR: Impossible error:", err.Error())
			http.Error(w, "", http.StatusInternalServerError)
			return

		}

		// write out the content type
		w.Header().Set("Cache-control", fmt.Sprintf("public, max-age=%d", int64(imageExpiry.Seconds())))
		w.Header().Set("Content-Type", *resp.ContentType)
		w.Header().Set("Etag", *resp.ETag)

		if err := waltz.Do(resp.Body, w, nil, resizeX, resizeY); err != nil {
			log.Println("Completed resizeHandler: ERROR: Resize Error:", err.Error())
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
	}
}

func robotsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	fmt.Fprintf(w, "User-Agent: *\nDisallow: /")
}

func main() {
	// fetch config from environment
	awsRegion := os.Getenv("AWS_REGION")
	awsBucketName := os.Getenv("AWS_BUCKET")
	awsBucketPrefix := os.Getenv("AWS_BUCKET_PREFIX")

	// create the s3 client
	client := s3.New(nil)

	// create middleware
	n := negroni.Classic()

	// create mux
	mux := http.NewServeMux()

	resizePath := fmt.Sprintf("/%s/", awsBucketPrefix)

	ro := resizeOpts{
		bucket: awsBucketName,
		prefix: awsBucketPrefix,
		path:   resizePath,
		region: awsRegion,
		client: client,
	}

	resizeHandler := newResizeHandler(ro)

	log.Println("Serving images from", resizePath)

	mux.HandleFunc(resizePath, resizeHandler)
	mux.HandleFunc("/robots.txt", robotsHandler)

	// add mux to middleware stack
	n.UseHandler(mux)

	server := &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%s", os.Getenv("PORT")),
		Handler: n,
	}

	log.Printf("Starting server on 0.0.0.0:%s\n", os.Getenv("PORT"))

	// run using gracehttp
	gracehttp.Serve(server)
}
