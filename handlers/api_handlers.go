package handlers

import (
	"log"
	"os"
	"time"

	"encoding/json"
	"net/http"

	"github.com/pakesson/cfs-server-go/utils"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type DownloadResponse struct {
	Url string `json:"url"`
}

type UploadResponse struct {
	Key string `json:"key"`
	Url string `json:"url"`
}

func oopsie(msg string) {
	log.Fatal(msg)
	os.Exit(1)
}

func ApiDownloadHandler(w http.ResponseWriter, r *http.Request) {
	bucket := os.Getenv("S3_BUCKET")
	region := os.Getenv("AWS_REGION")

	key := r.URL.Query().Get("key")
	log.Printf("Download request for '%v'", key)

	svc := s3.New(session.New(&aws.Config{Region: aws.String(region)}))
	_, err := svc.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		log.Println("Failed to get object: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	req, _ := svc.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	url, err := req.Presign(15 * time.Minute)

	if err != nil {
		log.Println("Failed to sign request", err)
	}

	w.Header().Set("Content-Type", "application/json")
	response := DownloadResponse{Url: url}
	json.NewEncoder(w).Encode(response)
}

func ApiUploadHandler(w http.ResponseWriter, r *http.Request) {
	filename := r.URL.Query().Get("filename")
	log.Printf("Upload request for '%v'", filename)

	key := utils.Uuid4()
	bucket := os.Getenv("S3_BUCKET")
	region := os.Getenv("AWS_REGION")

	svc := s3.New(session.New(&aws.Config{Region: aws.String(region)}))
	req, _ := svc.PutObjectRequest(&s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		ACL:    aws.String("private"),
		Metadata: map[string]*string{
			"filename": aws.String(filename),
		},
	})
	url, err := req.Presign(15 * time.Minute)

	if err != nil {
		log.Println("Failed to sign request", err)
	}

	w.Header().Set("Content-Type", "application/json")
	response := UploadResponse{Key: key, Url: url}
	json.NewEncoder(w).Encode(response)
}
