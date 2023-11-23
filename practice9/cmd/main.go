package main

import (
	"awesomeProject/web/api"
	"context"
	"flag"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"go.mongodb.org/mongo-driver/mongo/options"
	"io"
	"log"
	"net/http"
	"os"
)

var client *mongo.Client
var fs *gridfs.Bucket

func main() {
	// Connect to MongoDB
	ctx := context.Background()
	clientOptions := options.Client().ApplyURI("mongodb://mongo:27017")
	var err error
	client, err = mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)
	name := flag.String("name", "name", "")
	// Initialize GridFS
	fs, err = gridfs.NewBucket(
		client.Database(*name),
	)
	if err != nil {
		log.Fatal(err)
	}
	port := flag.String("port", "8080", "port to run the server on")
	

	flag.Parse()

	// Открытие файла для логирования
	logFile, err := os.OpenFile("server.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Ошибка открытия файла логов: %v", err)
	}

	// Установка вывода логов в файл и консоль
	logger := log.New(io.MultiWriter(os.Stdout, logFile), "", log.LstdFlags)
	router := mux.NewRouter()

	// Регистрация обработчиков различных методов
	router.HandleFunc("/api/files", func(w http.ResponseWriter, r *http.Request) {
		api.GetFiles(w, r, client, *name)
	}).Methods("GET")
	router.HandleFunc("/api/files", func(w http.ResponseWriter, r *http.Request) {
		api.UploadFile(w, r, client, fs, *name)
	}).Methods("POST")
	router.HandleFunc("/api/files/{id}", func(w http.ResponseWriter, r *http.Request) {
		api.GetFile(w, r, client, fs, *name)
	}).Methods("GET")
	router.HandleFunc("/api/files/{id}", func(w http.ResponseWriter, r *http.Request) {
		api.UpdateFile(w, r, client, fs, *name)
	}).Methods("PUT")
	router.HandleFunc("/api/files/{id}", func(w http.ResponseWriter, r *http.Request) {
		api.DeleteFile(w, r, client, fs, *name)
	}).Methods("DELETE")
	router.HandleFunc("/api/files/{id}/info", func(w http.ResponseWriter, r *http.Request) {
		api.GetFileInfo(w, r, client, *name)
	}).Methods("GET")
	logger.Println("Server started on port " + *port)
	logger.Fatal(http.ListenAndServe(":"+*port, router))
}
