package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type userData struct {
	Nama string `json:"Nama" binding:"required"`
}

func main() {

	err := godotenv.Load()

	if err != nil {
		log.Fatal("error loading .env File")
		return
	}

	db_user := os.Getenv("DB_USER")
	db_password := os.Getenv("DB_PASSWORD")
	db_name := os.Getenv("DB_NAME")
	db_port := os.Getenv("DB_PORT")

	r := gin.Default()
	dsn := fmt.Sprintf("host=localhost port=%s user=%s password=%s dbname=%s sslmode=disable",
		db_port, db_user, db_password, db_name)
	db, err := sqlx.Connect("postgres", dsn)

	if err != nil {
		log.Fatal("Gagal menyambung ke database")
		return
	}
	defer db.Close()

	r.POST("/data", func(c *gin.Context) {
		var data *userData
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"Error": "Gagal menerjemahkan",
			})
			return
		}

		_, err := db.Exec("INSERT INTO users(nama) VALUES($1)", data.Nama)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "gagal menginsert haha",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Data berhasil ditambahkan"})

	})

	r.Run(":8080")
}
