package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

func MySQLConnect() *sql.DB {
	// Connect to database
	db, err := sql.Open("mysql", os.Getenv("DSN"))
	if err != nil {
		panic(err)
	}
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)
	return db
}

type Task struct {
	ID			string `json:"id"`
	Name		string `json:"name"`
	Description string `json:"description"`
	CompletedAt string `json:"completedAt"`
}

var tasks []Task = []Task{
	{Name: "Task Name", Description: "Task Description"},
}

func CreateTask(c *gin.Context) {
	db := MySQLConnect()
	defer db.Close()

	stmtIns, err := db.Prepare("INSERT INTO tasks VALUES( DEFAULT, ?, ?, ? )") // ? = placeholder
	if err != nil {
		panic(err.Error())
	}
	defer stmtIns.Close()

	var newTask Task
	if err := c.BindJSON(&newTask); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Malformed JSON",
		})
		return
	}

	result, err := stmtIns.Exec(newTask.Name, newTask.Description, nil)
	fmt.Println("result added to db:", result)

	c.JSON(http.StatusOK, gin.H{
		"message": "Ok",
		"data": newTask,
	})
}

func ReadTasks(c *gin.Context) {
	db := MySQLConnect()
	defer db.Close()

	rows, err := db.Query("SELECT id, name, description, completedAt FROM tasks ORDER BY id DESC")
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "error executing the query",
		})
		return
	}

	var tasks []Task = []Task{}
	var id, name, description, completedAt []byte

	for rows.Next() {	// Fetch rows
		var newTask Task
		
		err = rows.Scan(&id, &name, &description, &completedAt)	// Get RawBytes from data
		if err != nil {
			panic(err.Error())
		}

		newTask.ID = string(id)
		newTask.Name = string(name)
		newTask.Description = string(description)
		newTask.CompletedAt = string(completedAt)
		tasks = append(tasks, newTask)
	}
	if err = rows.Err(); err != nil {
		panic(err.Error())
	}

	c.JSON(http.StatusOK, gin.H{
		"data": tasks,
	})
}

func ReadTask(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id")); 
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Not a valid ID",
		})
		return
	}

	db := MySQLConnect()
	defer db.Close()

	query := fmt.Sprintf("SELECT id, name, description, completedAt FROM tasks WHERE id=%d", id)
	row, err := db.Query(query)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "error executing the query",
		})
		return
	}

	var newTask Task
	var qid, name, description, completedAt []byte
	
	row.Next()
	err = row.Scan(&qid, &name, &description, &completedAt)	// Get RawBytes from data
	if err != nil {
		panic(err.Error())
	}

	newTask.ID = string(qid)
	newTask.Name = string(name)
	newTask.Description = string(description)
	newTask.CompletedAt = string(completedAt)
	
	if err = row.Err(); err != nil {
		panic(err.Error())
	}

	c.JSON(http.StatusOK, gin.H{
		"data": newTask,
	})
}

func UpdateTask(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id")); 
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Not a valid ID",
		})
		return
	}

	var newTask Task
	if err := c.BindJSON(&newTask); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Malformed JSON",
		})
		return
	}

	db := MySQLConnect()
	defer db.Close()

	rows, err := db.Query(fmt.Sprintf("UPDATE tasks SET name='%s', description='%s' WHERE id=%d", newTask.Name, newTask.Description, id))
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "error executing the query",
		})
		fmt.Println(rows)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Ok",
	})
}

func DeleteTask(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id")); 
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Not a valid ID",
		})
		return
	}
	
	db := MySQLConnect()
	defer db.Close()

	rows, err := db.Query(fmt.Sprintf("DELETE FROM tasks WHERE id=%d", id))
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "error executing the query",
		})
		fmt.Println(rows)
		return
	}

	// deletedTask := tasks[id]

	// firstHalf := tasks[:id]
	// secondHalf := tasks[id + 1:]
	// tasks = append(firstHalf, secondHalf...)

	c.JSON(http.StatusOK, gin.H{
		"message": "Task Deleted",
	})
}

func main() {
	// Load environment variable
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Println("Error loading .env file")
	}

	app := gin.Default()
	app.POST("/tasks", CreateTask)
	app.GET("/tasks", ReadTasks)
	app.GET("/tasks/:id", ReadTask)
	app.PUT("/tasks/:id", UpdateTask)
	app.DELETE("/tasks/:id", DeleteTask)
	app.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}