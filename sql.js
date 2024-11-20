const mysql = require("mysql2");
const fs = require("fs");
const path = require("path");

// Konfigurasi koneksi ke database
const connection = mysql.createConnection({
  host: "localhost", // Ganti dengan host MySQL Anda
  user: "sammy", // Ganti dengan username MySQL Anda
  password: "password", // Ganti dengan password MySQL Anda
  database: "times", // Nama database yang sudah ada
});

// Path ke file SQL yang akan diimpor
const sqlFilePath = path.join(__dirname, "timses.sql"); // Sesuaikan dengan lokasi file SQL Anda

// Fungsi untuk mengeksekusi SQL
const executeSqlFile = (filePath) => {
  // Membaca file SQL
  fs.readFile(filePath, "utf8", (err, sql) => {
    if (err) {
      console.error("Error membaca file SQL:", err);
      return;
    }

    // Menjalankan query SQL dalam file
    connection.query(sql, (error, results) => {
      if (error) {
        console.error("Error saat menjalankan query SQL:", error);
        connection.end();
        return;
      }
      console.log("Data berhasil diimpor:", results);
      connection.end();
    });
  });
};

// Mengeksekusi SQL
executeSqlFile(sqlFilePath);

// Menangani error koneksi
connection.connect((err) => {
  if (err) {
    console.error("Gagal terhubung ke database:", err);
  } else {
    console.log("Terhubung ke database MySQL");
  }
});
