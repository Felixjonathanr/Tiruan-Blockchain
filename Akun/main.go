package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"akun/key"
	"auth"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type userData struct {
	Nama string `form:"nama" binding:"required"`
	Pass string `form:"pass" binding:"required"`
}

type request struct {
	Username  string `json:"username"`
	Nama_user string `json:"nama_user"`
}

func pengecekanData(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}

func daftar(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var data userData
		if err := c.ShouldBind(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"Error": "Gagal menerjemahkan",
			})
			return
		}
		tx, err := db.Beginx()

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"PESAN": "Gagal menyambung ke transaksi",
			})
			return
		}
		defer tx.Rollback()

		_, err = tx.Exec("INSERT INTO users(nama,pass) VALUES($1,$2)", data.Nama, data.Pass)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "gagal menginsert haha",
			})
			return
		}
		kunciInggris, err := key.UserPrivateKey()
		if err != nil {
			panic(err)
		}

		_, err = tx.Exec("INSERT INTO pvkey(nama,pv_key) VALUES ($1,$2)", data.Nama, kunciInggris)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "ada error di server",
			})
			return
		}
		tx.Commit()

		c.JSON(http.StatusOK, gin.H{"message": "Data berhasil ditambahkan"})

	}
}

func login(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var dataInput userData
		var dataDB userData
		var pvKey string

		if err := c.ShouldBind(&dataInput); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"Error": "Internal server error, kami akan segera memperbaiki",
			})
			return
		}
		err := db.Get(&dataDB, "SELECT * FROM users WHERE nama=$1", dataInput.Nama)

		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"PESAN": "AKUN ANDA TIDAK DITEMUKAN, SEGERA LAKUKAN PENDAFTARAN",
			})
			return
		}
		if dataInput.Pass == dataDB.Pass {
			err := db.Get(&pvKey, "SELECT pv_key FROM pvkey WHERE nama=$1", dataInput.Nama)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Ada error di server kami, kalian tidak berkah masuk sini saatt!",
				})
				return
			}
			tokenSaya, err := auth.GenerateToken(dataInput.Nama, []byte(pvKey))

			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"Pesan": "Anda mau ngapain sat?",
				})
				return
			}
			c.Header("X-user-name", dataDB.Nama)           // Mengirim nama di header
			c.Header("Authorization", "Bearer "+tokenSaya) // Header standar

			c.JSON(http.StatusOK, gin.H{
				"Pesan": "Login Berhasil",
			})

			return
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"Pesan": "Password anda salah",
			})
			return
		}
	}
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

	r.POST("/daftar", daftar(db))
	r.POST("/login", login(db))
	r.POST("/cek/dataPembeli", pengecekanData(db))
	r.Run(":8080")
	fmt.Println("server berjalan di port 8080")

}
