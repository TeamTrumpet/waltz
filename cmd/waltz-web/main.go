package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/TeamTrumpet/waltz/waltz"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/codegangsta/negroni"
	"github.com/facebookgo/grace/gracehttp"
	"github.com/julienschmidt/httprouter"
)

const imageExpiry = 1 * time.Hour * 24 * 30

var (
	awsBucketName, awsRegion string
	client                   *s3.S3
)

func resizeHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	imageKey := ps.ByName("image_key")

	// fetch the query param
	resizeQuery := r.URL.Query().Get("resize")

	if resizeQuery == "" {
		// https://s3-us-west-2.amazonaws.com/armstrong-images/jellypus.png
		urlStr := fmt.Sprintf("https://s3-%s.amazonaws.com/%s/%s", awsRegion, awsBucketName, imageKey)
		http.Redirect(w, r, urlStr, http.StatusTemporaryRedirect)
		return
	}

	resizeX, resizeY, err := waltz.ParseResize(resizeQuery)
	if err != nil {
		fmt.Println("Request Error:", err.Error())
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	params := &s3.GetObjectInput{
		Bucket: aws.String(awsBucketName),
		Key:    aws.String(imageKey),
	}

	resp, err := client.GetObject(params)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {

			if reqErr, ok := err.(awserr.RequestFailure); ok {

				// A service error occurred
				fmt.Println("AWS Service error:", reqErr.Code(), reqErr.Message(), reqErr.StatusCode(), reqErr.RequestID())

				if reqErr.StatusCode() == http.StatusNotFound {
					http.Error(w, "", http.StatusNotFound)
					return
				}

				http.Error(w, "", http.StatusInternalServerError)
				return

			}

			// Generic AWS error with Code, Message, and original error (if any)
			fmt.Println("Generic AWS error:", awsErr.Code(), awsErr.Message(), awsErr.OrigErr())

			http.Error(w, "", http.StatusInternalServerError)
			return

		}

		// This case should never be hit, the SDK should always return an
		// error which satisfies the awserr.Error interface.
		fmt.Println("Impossible error:", err.Error())

		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	// write out the content type
	w.Header().Set("Content-Type", *resp.ContentType)

	if err := waltz.Do(resp.Body, w, nil, resizeX, resizeY); err != nil {
		fmt.Println("Resize Error:", err.Error())
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
}

func testHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, "<img src=\"/image/jellypus.png\"/>")

	for i := 500; i > 0; i -= 25 {
		fmt.Fprintf(w, "<img src=\"/image/jellypus.png?resize=%d\" width=\"%dpx\" height=\"%dpx\"/>", i, i, i)
	}
}

func main() {
	// fetch config from environment
	awsRegion = os.Getenv("AWS_REGION")
	awsBucketName = os.Getenv("AWS_BUCKET")

	// create the s3 client
	client = s3.New(nil)

	// create middleware
	n := negroni.Classic()

	// create mux
	mux := httprouter.New()

	// add routes
	mux.GET("/image/:image_key", resizeHandler)
	mux.GET("/test", testHandler)

	// add mux to middleware stack
	n.UseHandler(mux)

	server := &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%s", os.Getenv("PORT")),
		Handler: n,
	}

	// run using gracehttp
	gracehttp.Serve(server)
}
