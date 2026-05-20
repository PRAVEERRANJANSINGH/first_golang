package main

import (
	"fmt"
	"log"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/joho/godotenv"
	"os"
)

type Todo struct {
	ID       uint   `json:"id"`
	Body     string `json:"body"`
	Complete bool   `json:"complete"`
}

func main() {
	fmt.Println("Hello, World!")
err := godotenv.Load() //  Fixed
 if err != nil {
 	log.Fatal("Error loading .env file")
 }	
 PORT := os.Getenv("PORT")
if PORT == "" {
    PORT = "3000"
  }
	app := fiber.New()
	todos := []Todo{}

	app.Get("/api/todos", func(c fiber.Ctx) error {
		return c.Status(200).JSON(todos)
	})

	app.Post("/api/todos", func(c fiber.Ctx) error {
		todo := &Todo{}

		// Fiber v3 uses c.Bind().Body() instead of c.BodyParser()
		if err := c.Bind().Body(todo); err != nil {
			return err
		}

		if todo.Body == "" {
			return c.Status(400).JSON(fiber.Map{"msg": "body is required"})
		}

		todo.ID = uint(len(todos) + 1)
		todos = append(todos, *todo)

		return c.Status(201).JSON(todo)
	})

	// patch
	app.Patch("/api/todos/:id", func(c fiber.Ctx) error {
		// 1. Get the id parameter as a string from the URL
		idParam := c.Params("id")

		// 2. Convert that string into an integer (using standard library strconv)
		// We use Atoi (ASCII to Integer). It returns the number and an error if it fails.
		id, err := strconv.Atoi(idParam)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"msg": "Invalid ID format"})
		}

		// 3. Loop through your slice using the numeric ID
		for i, todo := range todos {
			// Typecast the converted 'id' to uint to match todo.ID
			if todo.ID == uint(id) {
				todos[i].Complete = true
				return c.Status(200).JSON(todos[i])
			}
		}

		// 4. Return 404 if no match was found
		return c.Status(404).JSON(fiber.Map{"msg": "todo not found"})
	})

	// delete
	app.Delete("/api/todos/:id", func(c fiber.Ctx) error {
		idParam := c.Params("id")
		id, err := strconv.Atoi(idParam)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"msg": "Invalid ID format"})
		}

		for i, todo := range todos {
			if todo.ID == uint(id) {
				todos = append(todos[:i], todos[i+1:]...)
				return c.Status(200).JSON(fiber.Map{"msg": "todo deleted"})
			}
		}

		return c.Status(404).JSON(fiber.Map{"msg": "todo not found"})
	})

	log.Fatal(app.Listen(":" + PORT))
}
