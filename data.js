const xlsx = require("xlsx");
const mysql = require("mysql2/promise");
const { v4: uuidv4 } = require("uuid"); // Menggunakan UUID v4

// Konfigurasi database
const dbConfig = {
  host: "localhost",
  user: "sammy",
  password: "password",
  database: "timses_data", // Ganti dengan nama database Anda
};

(async () => {
  try {
    // Buka koneksi database
    const connection = await mysql.createConnection(dbConfig);

    // Baca file Excel
    const workbook = xlsx.readFile("data.xlsx");
    const sheetName = workbook.SheetNames[0];
    const sheetData = xlsx.utils.sheet_to_json(workbook.Sheets[sheetName]);

    for (const row of sheetData) {
      const kecamatanName = row["Kecamatan"];
      const kelurahanName = row["Kelurahan"];
      const tpsName = row["TPS"];

      // Cari atau buat Kecamatan
      let [kecamatan] = await connection.execute(
        "SELECT * FROM indonesia_districts WHERE name = ?",
        [kecamatanName]
      );
      if (kecamatan.length === 0) {
        const [result] = await connection.execute(
          "INSERT INTO indonesia_districts (name, city_code) VALUES (?, ?)",
          [kecamatanName, 3202]
        );
        kecamatan = { id: result.insertId };
      } else {
        kecamatan = kecamatan[0];
      }

      // Cari atau buat Kelurahan
      let [kelurahan] = await connection.execute(
        "SELECT * FROM indonesia_villages WHERE name = ?",
        [kelurahanName]
      );
      if (kelurahan.length === 0) {
        const [result] = await connection.execute(
          "INSERT INTO indonesia_villages (name, district_code) VALUES (?, ?)",
          [kelurahanName, kecamatan.id]
        );
        kelurahan = { id: result.insertId };
      } else {
        kelurahan = kelurahan[0];
      }

      // Buat atau cari TPS dengan UUID
      await connection.execute(
        `
    INSERT INTO tps (id, name, village_id, district_id, lokasi, rt, rw, latitude, longitude, link_foto)
    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    ON DUPLICATE KEY UPDATE
        lokasi = VALUES(lokasi),
        rt = VALUES(rt),
        rw = VALUES(rw),
        latitude = VALUES(latitude),
        longitude = VALUES(longitude),
        link_foto = VALUES(link_foto)
  `,
        [
          uuidv4(), // Generate UUID baru
          `TPS ${tpsName}`, // Nama TPS
          kelurahan.id, // ID Kelurahan
          kecamatan.id, // ID Kecamatan
          row["Lokasi"] || null,
          row["RT"] || null,
          row["RW"] || null,
          row["Latitude"] || null,
          row["Longitude"] || null,
          row["Link Foto"] || null,
        ]
      );
    }

    console.log("Data berhasil diimpor!");
    connection.end();
  } catch (err) {
    console.error("Terjadi kesalahan:", err);
  }
})();
