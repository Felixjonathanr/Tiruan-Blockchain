package main

import (
	"bytes"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

/*
   id SERIAL PRIMARY KEY,
   nama_barang TEXT NOT NULL,
   jumlah INTEGER,
   nama_pembeli TEXT NOT NULL,
   nama_penjual TEXT NOT NULL,
   tandatangan_terakhir TEXT NOT NULL
*/

type dataPembelian struct {
	Nama_barang  string `form:"nama_barang" binding:"required"`
	Jumlah       int    `form:"jumlah" binding:"required"`
	Nama_penjual string `form:"nama_penjual" binding:"required"`
	Nama_penawar string
}

type request struct {
	Username  string `json:"username"`
	Nama_user string `json:"nama_user"`
}

type hasilRequest struct {
	Keterangan bool   `json:"keterangan"`
	Pv_key     string `json:"pvkey"`
}

var tandaTangan string

/*
if err := c.ShouldBind(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"Error": "Gagal menerjemahkan",
			})
			return
		}
*/

func decodeKey(key string) (*rsa.PrivateKey, error) {
	err := godotenv.Load()

	if err != nil {
		log.Fatal("error loading .env File")
		return nil, err
	}

	mk := []byte(os.Getenv("MASTER_KEY"))
	kuncibyte := []byte(key)

	block, _ := pem.Decode(kuncibyte)

	derBytes, err := x509.DecryptPEMBlock(block, mk)

	if err != nil {
		fmt.Println("GAGAL..... sakit hati lagi anjir lah")
		return nil, err
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(derBytes)

	if err != nil {
		fmt.Println("Kenapa lu harus hadir di hidup gua kalau engga jadi punyaku? hanying lah")
		return nil, err
	}

	return privateKey, nil

}

func pembeli(c *gin.Context) {
	var dataPembeli dataPembelian
	var hasil hasilRequest
	if err := c.ShouldBind(&dataPembeli); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"Error": "gagal menerjemahkan",
		})
		return
	}

	dataPembeli.Nama_penawar = c.GetHeader("X-user-name")

	reqData := request{
		Username:  dataPembeli.Nama_penjual,
		Nama_user: dataPembeli.Nama_penawar,
	}

	jsonData, err := json.Marshal(reqData)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"Error": "Gagal memverifikasi data pembeli",
		})
	}

	urlAkun := "http://localhost:8080/cek/dataPembeli"

	resp, err := http.Post(urlAkun, "application/json", bytes.NewBuffer(jsonData))

	if err != nil {
		fmt.Println("gagal mengirimkan url ke service b")
		return
	}

	defer resp.Body.Close()

	// mengambil data

	if err = json.NewDecoder(resp.Body).Decode(&hasil); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "pokoknya erro, anda tidak perlu tahu",
		})
		return
	}

	if !hasil.Keterangan {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Tujuannya siapa bangsattt!!! masukin user yang udah daftar, lu olang pengen bikin aplikasi aing nge hang?!",
		})
		return
	}

	kunciSaya, err := decodeKey(hasil.Pv_key)

	if err != nil {
		fmt.Println("Tuhan, jika dia bukan buat aku, tolong berikan yang sama persis sepertinya")
		return
	}

}

func penjual() {

}

func bank()

func main() {

}
