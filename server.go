package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type todo struct {
	ID      primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Content string             `json:"content,omitempty" bson:"content,omitempty"`
}

var (
	client *mongo.Client
	db     string = "Todo"
)

func getTodos(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	var todos []todo

	collection := client.Database(db).Collection("todos")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message": "` + err.Error() + `"}`))
		return
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var t todo
		cursor.Decode(&t)
		todos = append(todos, t)
	}

	if err := cursor.Err(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message": "` + err.Error() + `"}`))
		return
	}

	json.NewEncoder(w).Encode(todos)
}

func createTodo(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	var t todo

	json.NewDecoder(req.Body).Decode(&t)

	collection := client.Database(db).Collection("todos")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

	result, _ := collection.InsertOne(ctx, t)

	json.NewEncoder(w).Encode(result)
}

func deleteTodo(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	todoID, _ := primitive.ObjectIDFromHex(chi.URLParam(req, "id"))

	collection := client.Database(db).Collection("todos")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

	result, err := collection.DeleteOne(ctx, todo{ID: todoID})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message": "` + err.Error() + `"}`))
		return
	}

	json.NewEncoder(w).Encode(result)
}

func main() {
	fmt.Println("Starting server...")

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client, _ = mongo.NewClient(options.Client().ApplyURI("{Your MongoDB URI}"))
	_ = client.Connect(ctx)

	router := chi.NewRouter()

	router.Use(middleware.Logger)
	router.Use(useCors)

	router.Get("/todos", getTodos)
	router.Post("/todos", createTodo)
	router.Delete("/todos/{id}", deleteTodo)

	http.ListenAndServe(":5000", router)

}

// CORS middleware
func useCors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		(w).Header().Set("Access-Control-Allow-Origin", "*")
		(w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")

		next.ServeHTTP(w, req)
	})
}
