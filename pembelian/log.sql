CREATE TABLE IF NOT EXISTS catatan (
    id SERIAL PRIMARY KEY,
    nama_barang TEXT NOT NULL,
    jumlah INTEGER, 
    nama_pembeli TEXT NOT NULL,
    nama_penjual TEXT NOT NULL,
    tandatangan_terakhir TEXT NOT NULL
)

CREATE INDEX idx_catatan_nama_penjual ON catatan(nama_penjual)