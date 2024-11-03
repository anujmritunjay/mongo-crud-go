package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/anujmritunjay/mongo-crud-go/models"
	"github.com/julienschmidt/httprouter"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"gopkg.in/mgo.v2/bson"
)

type UserController struct {
	client *mongo.Client
}

func NewUserController(client *mongo.Client) *UserController {
	return &UserController{client: client}
}

func (uc UserController) GetUser(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	id := p.ByName("id")

	// Validate the ObjectID
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest) // Bad Request for invalid ID format
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Invalid ID format"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := uc.client.Database("your_database_name").Collection("users")

	// Find the user by ID
	var user models.User
	err = collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			w.WriteHeader(http.StatusNotFound) // Not Found if user doesn't exist
			json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "User not found"})
		} else {
			w.WriteHeader(http.StatusInternalServerError) // Internal Server Error for other issues
			json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": err.Error()})
		}
		return
	}

	// Successful retrieval
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "user": user})
}

func (uc *UserController) CreateUser(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
		return
	}

	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid JSON input: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Validate user data (add your validation rules)
	if err := validateUser(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Define a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Insert the user data into MongoDB
	collection := uc.client.Database("your_database_name").Collection("users")
	res, err := collection.InsertOne(ctx, user) // Use the user struct directly
	if err != nil {
		log.Printf("Failed to create user: %v", err)
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	// Create response
	response := map[string]interface{}{
		"success": true,
		"data":    res.InsertedID,
	}

	// Set headers
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	// Write JSON response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode response: %v", err)
		http.Error(w, "Failed to generate response", http.StatusInternalServerError)
		return
	}
}

func (uc *UserController) DeleteUser(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	id := p.ByName("id")

	objID, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		w.WriteHeader(500)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Invalid Object Id Provided"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	collection := uc.client.Database("your_database_name").Collection("users")

	result, err := collection.DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError) // Internal Server Error for other issues
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": err.Error()})
		return
	}
	if result.DeletedCount == 0 {
		w.WriteHeader(http.StatusNotFound) // Not Found if no documents matched
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "User not found"})
		return
	}

	// Successful deletion
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "message": "User deleted successfully"})

}

// Helper function to validate user data
func validateUser(user *models.User) error {
	if user.Name == "" {
		return fmt.Errorf("name is required")
	}
	if user.Age <= 0 {
		return fmt.Errorf("age must be positive")
	}
	return nil
}
