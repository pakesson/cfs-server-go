package handlers

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"

	"encoding/base64"
	"html/template"
	"io/ioutil"
	"net/http"

	"crypto/rand"
	"crypto/sha256"

	"github.com/pakesson/cfs-server-go/utils"

	"golang.org/x/crypto/nacl/secretbox"

	"github.com/gorilla/mux"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

const (
	KEY_SIZE   = 32
	NONCE_SIZE = 24
	MAX_MEMORY = 10 * 1024 * 1024

	TITLE = "CCFCSC - Cloud Crypto File Cloud Storage for the Cloud(tm)"
)

func createNonce() (*[NONCE_SIZE]byte, error) {
	nonce := new([NONCE_SIZE]byte)
	_, err := io.ReadFull(rand.Reader, nonce[:])
	if err != nil {
		return nil, err
	}
	return nonce, nil
}

func WebUploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	template_data := make(map[string]string)
	template_data["title"] = TITLE

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmpl.Execute(w, template_data)
}

func WebUploadFileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	bucket := os.Getenv("S3_BUCKET")
	region := os.Getenv("AWS_REGION")

	// Data exceeding MAX_MEMORY will be stored in temporary files
	if err := r.ParseMultipartForm(MAX_MEMORY); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	defer r.MultipartForm.RemoveAll() // Make sure temporary files (if any) are removed

	password := r.FormValue("password")
	if password == "" {
		log.Println("Empty password")
		http.Error(w, "Empty password", http.StatusBadRequest)
		return
	}
	cipher_key := sha256.Sum256([]byte(password))

	file, header, err := r.FormFile("uploadFile")
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer file.Close()

	nonce, err := createNonce()
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	filename := header.Filename
	encrypted_filename := make([]byte, len(nonce))
	copy(encrypted_filename, nonce[:])
	encrypted_filename = secretbox.Seal(encrypted_filename, []byte(filename), nonce, &cipher_key)
	encoded_filename := base64.StdEncoding.EncodeToString(encrypted_filename)

	data, err := ioutil.ReadAll(file)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	nonce, err = createNonce()
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	encrypted_data := make([]byte, len(nonce))
	copy(encrypted_data, nonce[:])
	encrypted_data = secretbox.Seal(encrypted_data, data, nonce, &cipher_key)

	key := utils.Uuid4()

	svc := s3.New(session.New(&aws.Config{Region: aws.String(region)}))
	_, err = svc.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		ACL:    aws.String("private"),
		Body:   bytes.NewReader(encrypted_data),
		Metadata: map[string]*string{
			"filename": aws.String(encoded_filename),
		},
	})
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	template_data := make(map[string]interface{})
	template_data["title"] = TITLE
	template_data["message"] = template.HTML(fmt.Sprintf("File uploaded. <a href=\"http://localhost:5000/download/%s\">http://localhost:5000/download/%s</a>.", key, key))

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmpl.Execute(w, template_data)
}

func WebDownloadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	bucket := os.Getenv("S3_BUCKET")
	region := os.Getenv("AWS_REGION")

	vars := mux.Vars(r)
	key := vars["key"]

	svc := s3.New(session.New(&aws.Config{Region: aws.String(region)}))
	_, err := svc.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		log.Println("Failed to get object: ", err)
		http.NotFound(w, r)
		return
	}

	tmpl, err := template.ParseFiles("templates/download.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	template_data := make(map[string]string)
	template_data["title"] = TITLE
	template_data["key"] = key

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmpl.Execute(w, template_data)
}

func WebDownloadFileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	vars := mux.Vars(r)
	key := vars["key"]

	password := r.FormValue("password")
	if password == "" {
		log.Println("Empty password")
		http.Error(w, "Empty password", http.StatusBadRequest)
		return
	}
	cipherKey := sha256.Sum256([]byte(password))

	bucket := os.Getenv("S3_BUCKET")
	region := os.Getenv("AWS_REGION")

	svc := s3.New(session.New(&aws.Config{Region: aws.String(region)}))
	resp, err := svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer resp.Body.Close()

	// Metadata keys are always returned in uppercase due to a bug
	// https://github.com/aws/aws-sdk-go/issues/445
	encodedFilename := resp.Metadata["Filename"]
	if encodedFilename == nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	encryptedFilename, err := base64.StdEncoding.DecodeString(*encodedFilename)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var nonce [NONCE_SIZE]byte
	copy(nonce[:], encryptedFilename)
	filenameBytes, ok := secretbox.Open([]byte{}, encryptedFilename[NONCE_SIZE:], &nonce, &cipherKey)
	if !ok { // Thanks for returning a bool instead of an error object :( Super intuitive!
		http.Error(w, "Incorrect password", http.StatusBadRequest)
		return
	}
	filename := string(filenameBytes)

	encryptedBuffer, err := ioutil.ReadAll(resp.Body)
	copy(nonce[:], encryptedBuffer)
	buffer, ok := secretbox.Open([]byte{}, encryptedBuffer[NONCE_SIZE:], &nonce, &cipherKey)
	if !ok {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	disposition := fmt.Sprintf("attachment; filename=\"%s\"", filename)
	w.Header().Set("Content-Disposition", disposition)
	w.Header().Set("Content-Length", strconv.Itoa(len(buffer)))

	w.Write(buffer)
}
