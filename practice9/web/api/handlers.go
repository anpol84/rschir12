package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type File struct {
	ID          primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Name        string             `json:"name,omitempty" bson:"name,omitempty"`
	Description string             `json:"description,omitempty" bson:"description,omitempty"`
	Size        int64              `json:"size,omitempty" bson:"size,omitempty"`
}

var logFile, _ = os.OpenFile("server.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)

// Установка вывода логов в файл и консоль
var logger = log.New(io.MultiWriter(os.Stdout, logFile), "", log.LstdFlags)

func GetFiles(w http.ResponseWriter, r *http.Request, client *mongo.Client, name string) {

	// Get all files from MongoDB
	ctx := context.Background()
	collection := client.Database(name).Collection(name)
	cur, err := collection.Find(ctx, bson.M{})
	if err != nil {
		log.Println(err)
		http.Error(w, "Error getting files", http.StatusInternalServerError)
		return
	}
	defer cur.Close(ctx)

	// Convert to slice of File structs and encode as JSON
	var files []File
	for cur.Next(ctx) {
		var file File
		err := cur.Decode(&file)
		if err != nil {
			log.Println(err)
			http.Error(w, "Error decoding files", http.StatusInternalServerError)
			return
		}
		files = append(files, file)
	}
	err = cur.Err()
	if err != nil {
		log.Println(err)
		http.Error(w, "Error iterating files", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(files)
	logger.Println("Files was got")
}

func GetFile(w http.ResponseWriter, r *http.Request, client *mongo.Client, fs *gridfs.Bucket, name string) {

	// Get file ID from URL parameter
	vars := mux.Vars(r)
	id, err := primitive.ObjectIDFromHex(vars["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	// Get file from MongoDB
	ctx := context.Background()
	collection := client.Database(name).Collection(name)
	filter := bson.M{"_id": id}
	var file File
	err = collection.FindOne(ctx, filter).Decode(&file)
	if err != nil {
		log.Println(err)
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	// Get file from GridFS
	fileBytes, err := getFileBytes(id, fs)
	if err != nil {
		log.Println(err)
		http.Error(w, "Error getting file", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", file.Name))
	w.Write(fileBytes)
	logger.Println("File was read")
}

func GetFileInfo(w http.ResponseWriter, r *http.Request, client *mongo.Client, name string) {

	// Get file ID from URL parameter
	vars := mux.Vars(r)
	id, err := primitive.ObjectIDFromHex(vars["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	// Get file from MongoDB
	ctx := context.Background()
	collection := client.Database(name).Collection(name)
	filter := bson.M{"_id": id}
	var file File
	err = collection.FindOne(ctx, filter).Decode(&file)
	if err != nil {
		log.Println(err)
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fileInfo := map[string]interface{}{
		"id":          file.ID.Hex(),
		"name":        file.Name,
		"description": file.Description,
		"size":        file.Size,
	}
	json.NewEncoder(w).Encode(fileInfo)
	logger.Println("File info was read")
}

func UploadFile(w http.ResponseWriter, r *http.Request, client *mongo.Client, fs *gridfs.Bucket, name string) {

	file, handler, err := r.FormFile("file")
	if err != nil {
		log.Println(err)
		http.Error(w, "Error reading file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// Create file document in MongoDB
	ctx := context.Background()
	collection := client.Database(name).Collection(name)
	fileSize := handler.Size
	result, err := collection.InsertOne(ctx, bson.M{
		"name":        r.FormValue("name"),
		"contentType": handler.Header.Get("Content-Type"),
		"description": r.FormValue("description"),
		"size":        fileSize,
	})
	if err != nil {
		log.Println(err)
		http.Error(w, "Error inserting file", http.StatusInternalServerError)
		return
	}
	id := result.InsertedID.(primitive.ObjectID)

	// Upload file to GridFS
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		log.Println(err)
		http.Error(w, "Error reading file", http.StatusInternalServerError)
		return
	}
	err = uploadFileBytes(id, fileBytes, fs)
	if err != nil {
		log.Println(err)
		http.Error(w, "Error uploading file", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	logger.Println("File was created")
}

func UpdateFile(w http.ResponseWriter, r *http.Request, client *mongo.Client, fs *gridfs.Bucket, name string) {

	vars := mux.Vars(r)
	id, err := primitive.ObjectIDFromHex(vars["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	file, handler, err := r.FormFile("file")
	if err != nil {
		log.Println(err)
		http.Error(w, "Error reading file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// Update file document in MongoDB
	ctx := context.Background()
	collection := client.Database(name).Collection(name)
	filter := bson.M{"_id": id}
	fileSize := handler.Size
	update := bson.M{
		"$set": bson.M{
			"name":        r.FormValue("name"),
			"contentType": handler.Header.Get("Content-Type"),
			"description": r.FormValue("description"),
			"size":        fileSize,
		},
	}
	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		log.Println(err)
		http.Error(w, "Error updating file", http.StatusInternalServerError)
		return
	}
	if result.ModifiedCount == 0 {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	// Upload updated file to GridFS
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		log.Println(err)
		http.Error(w, "Error reading file", http.StatusInternalServerError)
		return
	}
	err = fs.Delete(id)

	err = uploadFileBytes(id, fileBytes, fs)
	if err != nil {
		log.Println(err)
		http.Error(w, "Error uploading file", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
	logger.Println("File was updated")
}

func DeleteFile(w http.ResponseWriter, r *http.Request, client *mongo.Client, fs *gridfs.Bucket, name string) {

	// Get file ID from URL parameter
	vars := mux.Vars(r)
	id, err := primitive.ObjectIDFromHex(vars["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	// Delete file document from MongoDB
	ctx := context.Background()
	collection := client.Database(name).Collection(name)
	filter := bson.M{"_id": id}
	result, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		log.Println(err)
		http.Error(w, "Error deleting file", http.StatusInternalServerError)
		return
	}
	if result.DeletedCount == 0 {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	// Delete file from GridFS
	err = fs.Delete(id)
	if err != nil {
		log.Println(err)
		http.Error(w, "Error deleting file", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
	logger.Println("File was deleted")
}

func getFileBytes(id primitive.ObjectID, fs *gridfs.Bucket) ([]byte, error) {
	downloadStream, err := fs.OpenDownloadStream(id)
	if err != nil {
		return nil, err
	}
	defer downloadStream.Close()

	fileBytes, err := ioutil.ReadAll(downloadStream)
	if err != nil {
		return nil, err
	}

	return fileBytes, nil
}

func uploadFileBytes(id primitive.ObjectID, fileBytes []byte, fs *gridfs.Bucket) error {

	// Удаление предыдущего содержимого файла

	uploadStream, err := fs.OpenUploadStreamWithID(id, id.Hex())
	if err != nil {
		return err
	}
	defer uploadStream.Close()

	_, err = uploadStream.Write(fileBytes)
	if err != nil {
		return err
	}

	return nil
}
