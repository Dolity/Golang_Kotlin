package handler

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"
	_ "github.com/lib/pq"
)

const (
	dbUser     = "postgres"
	dbPassword = "abc123"
	dbName     = "todos"
	dbHost     = "localhost"
	dbPort     = "5432"
)

type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func connectDB() (*sql.DB, error) {
	connectionString :=
		fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%s sslmode=disable",
			dbUser, dbPassword, dbName, dbHost, dbPort)
	return sql.Open("postgres", connectionString)
}

func main() {
	db, err := connectDB()
	if err != nil {
		log.Fatal("Could not connect to the database:", err)
	}
	defer db.Close()

	app := fiber.New()

	// Routes
	app.Get("/users", getUsersHandler(db))
	app.Post("/users", createUserHandler(db))
	app.Put("/users/:id", updateUserHandler(db))
	app.Delete("/users/:id", deleteUserHandler(db))

	app.Post("/login", loginHandler(db))

	// Start the server
	// app.Listen(":8080")
	// app.Listen("127.0.0.1:8080")
	// log.Fatal(app.Listen(":3000"))
	log.Fatal(app.ListenTLS(":3000", "server.crt", "server.key"))

}

// Handler functions

func getUsersHandler(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		users, err := getUsers(db)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not get users"})
		}
		return c.JSON(users)
	}
}

func createUserHandler(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var user User
		if err := c.BodyParser(&user); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
		}

		id, err := createUser(db, user.Name, user.Age)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not create user"})
		}
		user.ID = id
		return c.JSON(user)
	}
}

func updateUserHandler(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		idStr := c.Params("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
		}

		var user User
		if err := c.BodyParser(&user); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
		}
		user.ID = id

		err = updateUser(db, user)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not update user"})
		}
		return c.JSON(user)
	}
}

func deleteUserHandler(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")

		err := deleteUser(db, id)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not delete user"})
		}
		return c.SendStatus(fiber.StatusNoContent)
	}
}

// CRUD functions
func getUsers(db *sql.DB) ([]User, error) {
	query := "SELECT id, name, age FROM person"
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Name, &user.Age); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

func createUser(db *sql.DB, name string, age int) (int, error) {
	query := "INSERT INTO person (name, age) VALUES ($1, $2) RETURNING id"
	var id int
	err := db.QueryRow(query, name, age).Scan(&id)
	return id, err
}

func updateUser(db *sql.DB, user User) error {
	query := "UPDATE person SET name=$1, age=$2 WHERE id=$3"
	_, err := db.Exec(query, user.Name, user.Age, user.ID)
	return err
}

func deleteUser(db *sql.DB, id string) error {
	query := "DELETE FROM person WHERE id=$1"
	_, err := db.Exec(query, id)
	return err
}

// //////////////////////////////////////////////////////////////////////////////////////////////////////////
func loginHandler(db *sql.DB) fiber.Handler {
	type LoginRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	type LoginResponse struct {
		User User `json:"user"`
	}

	return func(c *fiber.Ctx) error {
		var req LoginRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
		}

		// ตรวจสอบ username และ password ในตาราง role
		query := "SELECT username, password FROM role WHERE username = $1 AND password = $2"
		var dbUsername, dbPassword string
		err := db.QueryRow(query, req.Username, req.Password).Scan(&dbUsername, &dbPassword)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid username or password"})
		}

		// หากตรวจสอบเรียบร้อยและ username/password ถูกต้อง ดึงข้อมูลจากตาราง person
		users, err := getUsers(db)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not get users"})
		}

		// ส่งข้อมูลของผู้ใช้ที่ล็อกอินแล้วกลับไป
		response := users
		return c.JSON(response)
	}
}

func Hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "<h1>Hello from Go!</h1>")
}
