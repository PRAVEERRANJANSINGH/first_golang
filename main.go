package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Todo struct {
	ID       primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Body     string             `json:"body" bson:"body"`
	Complete bool               `json:"complete" bson:"complete"`
}

var collection *mongo.Collection

func main() {
  fmt.Println("Starting server...")

  // 1. Load Environment Variables
  err := godotenv.Load(".env")
  if err != nil {
    log.Fatal("Error loading .env file: ", err)
  }

  PORT := os.Getenv("PORT")
  if PORT == "" {
    PORT = "3000"
  }
  MONGODB_URI := os.Getenv("MONGODB_URI")

  // 2. Setup MongoDB Connection with Timeout Context
  ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
  defer cancel()

  client, err := mongo.Connect(ctx, options.Client().ApplyURI(MONGODB_URI))
  if err != nil {
    log.Fatal("Error connecting to MongoDB: ", err)
  }

  // 3. Ping the database to verify connectivity
  err = client.Ping(ctx, nil)
  if err != nil {
    log.Fatal("Error pinging MongoDB: ", err)
  }

  fmt.Println("Connected to MongoDB successfully!")
  collection = client.Database("go-todo-app").Collection("todos")

  // 4. Initialize Fiber App & Routes
  app := fiber.New()

  app.Get("/api/todos", getTodos)
  app.Post("/api/todos", createTodo)
  app.Patch("/api/todos/:id", updateTodo) 
  app.Delete("/api/todos/:id", deleteTodo)

  log.Fatal(app.Listen(":" + PORT))
}
// 5. GET All Todos
func getTodos(c fiber.Ctx) error {
	var todos []Todo

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, bson.D{})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch todos from database"})
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &todos); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to parse database documents"})
	}

	if todos == nil {
		todos = []Todo{}
	}

	return c.Status(200).JSON(todos)
}

// 6. CREATE a Todo
func createTodo(c fiber.Ctx) error {
	todo := new(Todo)

	// Fiber v3 parsing syntax
	if err := c.Bind().Body(todo); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if todo.Body == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Todo body cannot be empty"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Insert into MongoDB
	result, err := collection.InsertOne(ctx, todo)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to insert todo"})
	}

	// Assign the automatically generated ObjectID back to our struct for response
	todo.ID = result.InsertedID.(primitive.ObjectID)

	return c.Status(201).JSON(todo)
}

// 7. UPDATE (Toggle complete) a Todo
func updateTodo(c fiber.Ctx) error {
	idParam := c.Params("id")

	// Convert string hex ID into a native MongoDB ObjectID
	objectID, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid ID format"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"_id": objectID}
	update := bson.M{"$set": bson.M{"complete": true}} // Force updates complete status to true

	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update todo"})
	}

	if result.MatchedCount == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "Todo not found"})
	}

	return c.Status(200).JSON(fiber.Map{"success": true, "msg": "Todo marked as complete"})
}

// 8. DELETE a Todo
func deleteTodo(c fiber.Ctx) error {
	idParam := c.Params("id")

	objectID, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid ID format"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"_id": objectID}
	result, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to delete todo"})
	}

	if result.DeletedCount == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "Todo not found"})
	}

	return c.Status(200).JSON(fiber.Map{"success": true, "msg": "Todo deleted successfully"})
}