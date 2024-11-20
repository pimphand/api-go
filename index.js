const express = require("express");
const mysql = require("mysql2");
const app = express();
const port = 3000; // Ganti dengan port yang Anda inginkan

// Konfigurasi koneksi ke database
const connection = mysql.createConnection({
  host: "153.92.15.21", // Ganti dengan host MySQL Anda
  user: "u268977163_asjab_new", // Ganti dengan username MySQL Anda
  password: "Asjap2024@", // Ganti dengan password MySQL Anda
  database: "u268977163_asjab_new", // Nama database yang sudah ada
});

// API untuk mendapatkan data TPS
app.get("/api/tps", (req, res) => {
  const query = `
    SELECT 
      COUNT(t.id) AS tps_count,
      SUM(CASE WHEN dr.id IS NOT NULL THEN 1 ELSE 0 END) AS tps_with_data,
      SUM(CASE WHEN dr.id IS NULL THEN 1 ELSE 0 END) AS tps_without_data
    FROM tps t
    LEFT JOIN data_recaps dr ON t.id = dr.tps_id;
  `;

  connection.query(query, (err, results) => {
    if (err) {
      console.error("Error querying the database:", err);
      return res.status(500).json({ error: "Error querying the database" });
    }

    // Data untuk grafik TPS
    const tpsData = {
      labels: ["Sudah Input", "Belum Input"],
      datasets: [
        {
          data: [results[0].tps_with_data, results[0].tps_without_data],
          backgroundColor: ["#10b981", "#ef4444"], // Warna untuk grafik
        },
      ],
    };

    // Kirim data grafik ke frontend
    res.json(tpsData);
  });
});

// API untuk mendapatkan detail kecamatan berdasarkan TPS, district_id, village_id, dan hasData
app.get("/api/tps/detail", (req, res) => {
  // Mendapatkan query params
  const districtId = req.query.districtId;
  const villageId = req.query.villageId;
  const hasData = req.query.hasData === "true" ? "IS NOT NULL" : "IS NULL";

  // Membuat query dasar
  let query = `
    SELECT 
      id, name
    FROM indonesia_districts
    WHERE id IN (
      SELECT DISTINCT district_id
      FROM tps
      LEFT JOIN data_recaps dr ON tps.id = dr.tps_id
      WHERE dr.id ${hasData}
    )
  `;

  // Menambahkan filter berdasarkan districtId jika ada
  if (districtId) {
    query += ` AND id = ?`;
  }

  // Menambahkan filter berdasarkan villageId jika ada
  if (villageId) {
    query += ` AND id IN (
        SELECT DISTINCT village_id
        FROM tps
        LEFT JOIN data_recaps dr ON tps.id = dr.tps_id
        WHERE dr.id ${hasData}
      )`;
  }

  // Jalankan query
  connection.query(
    query,
    [districtId, villageId].filter(Boolean),
    (err, results) => {
      if (err) {
        console.error("Error querying the database:", err);
        return res.status(500).json({ error: "Error querying the database" });
      }

      // Kirim data kecamatan ke frontend
      res.json({
        districts: results,
      });
    }
  );
});

// Start server
app.listen(port, () => {
  console.log(`Server is running on http://localhost:${port}`);
});
