package main

import (
	"akun/key"
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/skip2/go-qrcode"

	"github.com/gin-gonic/gin"
)

/*
   id SERIAL PRIMARY KEY,
   nama_barang TEXT NOT NULL,
   jumlah INTEGER,
   nama_pembeli TEXT NOT NULL,
   nama_penjual TEXT NOT NULL,
   tandatangan_terakhir TEXT NOT NULL
*/

// untuk menerima input dari pengguna tentang transaksi
type dataPembelian struct {
	Nama_barang  string `form:"nama_barang" binding:"required" json:"nama_barang"`
	Jumlah       int    `form:"jumlah" binding:"required" json:"jumlah"`
	Nama_penjual string `form:"nama_penjual" binding:"required" json:"nama_penjual"`
	Nama_penawar string `json:"nama_penawar"`
}

// untuk verifikasi apakah nama penjual itu ada di dalam database
type request struct {
	Username  string `json:"username"`
	Nama_user string `json:"nama_user"`
}

// menampung hasil request, karena untuk verifikasi harus melalui microservice akun
type hasilRequest struct {
	Keterangan bool   `json:"keterangan"`
	Pv_key     string `json:"pvkey"`
}

type kirimKeaksiPenjual struct {
	StringKeseluruhan    string        `json:"datakeseluruhan" `
	Keputusan            string        `json:"keputusan" form:"keputusan"`
	DataTransaksiPenawar dataPembelian `json:"data_pembelian"`
}


func signed(data dataPembelian, pvKey string) (string, error) {
	// kuncinya pembeli, bukan kunci penjual
	kunciSaya, err := key.DecodeKey(pvKey)

	if err != nil {
		fmt.Println("Tuhan, jika dia bukan buat aku, tolong berikan yang sama persis sepertinya")
		return "", nil
	}

	// di sini mekanisme signed data

	dataJson, err := json.Marshal(data)

	if err != nil {
		fmt.Println("Ada error di bagian pembuatan signed data")
		return "", nil
	}

	hash := sha256.Sum256(dataJson)

	signature, err := rsa.SignPKCS1v15(rand.Reader, kunciSaya, crypto.SHA256, hash[:])

	if err != nil {
		fmt.Println("Gagal membuat signature")
		return "", nil
	}

	dataAkhir := fmt.Sprintf("%s.%s",
		base64.StdEncoding.EncodeToString(dataJson),
		base64.StdEncoding.EncodeToString(signature),
	)

	return dataAkhir, nil
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
	// hasil.Pv_key itu untuk pembeli

	// di sini mekanisme signed data

	dataAkhir, err := signed(dataPembeli, hasil.Pv_key)

	png, err := qrcode.Encode(dataAkhir, qrcode.Medium, 256)

	if err != nil {
		fmt.Println("Hita nadua gabe sada")
		return
	}

	c.Data(200, "image/png", png)

}

func penjual(c *gin.Context) {

	var stringPenawar string
	var dataKiriman string
	var tandaPenawar string
	var dataTransaksi dataPembelian


	// logika scan barcode
	// dia akan mengambil string berisi data dan signature

	// untuk memisahkan data dan tandatangan penawar
	_, err := fmt.Sscanf(stringPenawar, "%s.%s", &dataKiriman, &tandaPenawar)

	if err != nil {
		fmt.Println("Gagal menerjemahkan data")
		return
	}

	// decode string data
	terjemahanData, err := base64.StdEncoding.DecodeString(dataKiriman)

	if err != nil {
		fmt.Println("gagal mengundecode data")
		return
	}

	// mengembalikan data yang dikirim ke variabel data transaksi
	if err := json.Unmarshal(terjemahanData, &dataTransaksi); err != nil {
		fmt.Println("Gagal parse json")
		return
	}

	// mengirim data ke html pengguna
	c.JSON(http.StatusAccepted, dataPembelian{
		Nama_barang:  dataTransaksi.Nama_barang,
		Jumlah:       dataTransaksi.Jumlah,
		Nama_penjual: dataTransaksi.Nama_penjual,
		Nama_penawar: dataTransaksi.Nama_penawar,
	})
	if err := c.ShouldBind(&keputusan)
	dataTransaksiGacor := dataPembelian{
		Nama_barang:  dataTransaksi.Nama_barang,
		Jumlah:       dataTransaksi.Jumlah,
		Nama_penjual: dataTransaksi.Nama_penjual,
		Nama_penawar: dataTransaksi.Nama_penawar,
	}

	data := kirimKeaksiPenjual{
		StringKeseluruhan:    tandaPenawar,
		DataTransaksiPenawar: dataTransaksiGacor,
	}

	jesonData, err := json.Marshal(data)

	if err != nil {
		fmt.Println("Surti tejo")
		return
	}

	http.Post("/aksiPenjual", "application/json", bytes.NewBuffer(jesonData))

}

func aksiPenjual(c *gin.Context) {

	// ini untuk ngambil string awal, biar diakhir ada tiga signedstring

	var hasilPv_key hasilRequest

	var dataAseli kirimKeaksiPenjual

	if err := c.ShouldBindJSON(&dataAseli); err != nil {
		fmt.Println("aku jatuh cinta ")
		return
	}
	// buat ngambil string keseluruhan yang berisikan data dan

	if dataAseli.Keputusan == "tidak" {
		c.Redirect(http.StatusFound, "/indeks.html")
		return
	}

	// karena nanti dia akses dari loginnya dia, harusnya masih ada namanya di header, jadi kita tinggal ambil dari header, trus minta dari akun

	if dataAseli.Keputusan == "setuju" {
		namaSaya := c.GetHeader("X-user-name")

		reqData := request{
			Username:  namaSaya,
			Nama_user: namaSaya,
		}

		jsonData, err := json.Marshal(reqData)
		if err != nil {
			fmt.Println("ku tak punya pilihan yang dikendali pikiran")
			return
		}

		urlAkun := "http://localhost:8080/cek/dataPembeli"

		resp, err := http.Post(urlAkun, "application/json", bytes.NewBuffer(jsonData))

		if err != nil {
			fmt.Println("Sedia aku sebelum hujan, apa yang kau butuh ku berikan")
			return
		}

		defer resp.Body.Close()

		if err := json.NewDecoder(resp.Body).Decode(&hasilPv_key); err != nil {
			fmt.Println("Aku tak tau, mengapa aku begitu, setiap engkau di sini, hidupkan langit jiwaku")
			return
		}

		dataAkhir, err := signed(dataAseli.DataTransaksiPenawar, hasilPv_key.Pv_key)

		if err != nil {
			fmt.Println("kacauuu, kapan muncul titanic 2 sih?!!!")
			return
		}

		dataAkhirJSON, err := json.Marshal(dataAkhir)

		http.Post("/bank", "application/json", bytes.NewBuffer(dataAkhirJSON))
	}

}

func bank() {
	// logika pengambilan dan signed bank
}

func main() {

	r := gin.Default()

	r.Run(":8081")

}
