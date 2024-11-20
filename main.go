package main

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Model untuk TPS
type TPS struct {
	ID   uint `gorm:"primaryKey"`
	Name string
}

// Model untuk DataRecap
type DataRecap struct {
	ID   uint `gorm:"primaryKey"`
	TpsID uint
}


// Variabel global untuk GORM
var db *gorm.DB

// Fungsi untuk inisialisasi koneksi ke database dengan GORM
func initDB() {
	var err error
	// Ganti dengan kredensial yang sesuai
	dsn := "u268977163_asjab_new:Asjap2024@@tcp(153.92.15.21:3306)/u268977163_asjab_new?charset=utf8mb4&parseTime=True&loc=Local"
	// Inisialisasi logger untuk mencatat query
	newLogger := logger.New(
		log.New(log.Writer(), "\r\n", log.LstdFlags), // Log writer
		logger.Config{
			SlowThreshold: 200 * time.Millisecond, // Set threshold for slow queries
			LogLevel:      logger.Info,            // Log level
			Colorful:      true,                   // Enable colorful output
		},
	)
	// Menghubungkan ke database dengan GORM dan logger
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		log.Fatal("Error opening database connection:", err)
	}
}
type TPSData struct {
    TpsName        string `gorm:"column:tps_name"`
    TpsCount       int64  `gorm:"column:tpsCount"`
    TpsWithData    int64  `gorm:"column:tpsWithData"`
    TpsWithoutData int64  `gorm:"column:tpsWithoutData"`
}
// API untuk mendapatkan data TPS
func getTPSData(c *gin.Context) {
	var tpsData TPSData

	// Query untuk menghitung TPS dengan data dan tanpa data
	err := db.Table("tps").
		Select("COUNT(tps.id) AS tps_count, "+
			"SUM(CASE WHEN data_recaps.id IS NOT NULL THEN 1 ELSE 0 END) AS tps_with_data, "+
			"SUM(CASE WHEN data_recaps.id IS NULL THEN 1 ELSE 0 END) AS tps_without_data").
		Joins("LEFT JOIN data_recaps ON tps.id = data_recaps.tps_id").
		Scan(&tpsData).Error

	// Log error jika ada
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error querying the database"})
		log.Println("Error querying the database:", err)
		return
	}

	// Data untuk grafik TPS
	tpsResponse := map[string]interface{}{
		"labels": []string{"Sudah Input", "Belum Input"},
		"datasets": []map[string]interface{}{
			{
				"data":           []int{int(tpsData.TpsWithData), int(tpsData.TpsWithoutData)},
				"backgroundColor": []string{"#10b981", "#ef4444"},
			},
		},
	}

	// Kirim data grafik ke frontend
	c.JSON(http.StatusOK, tpsResponse)
}

// API untuk mendapatkan detail kecamatan berdasarkan TPS, district_id, village_id, dan hasData
type District struct {
    ID          int    `gorm:"column:id"`
    Name        string `gorm:"column:name"`
    TpsCount    int64  `gorm:"column:tpsCount"`
    TpsWithData int64  `gorm:"column:tpsWithData"`
    TpsWithoutData int64 `gorm:"column:tpsWithoutData"`
}

func getDistrict(c *gin.Context) {
	districtId := c.DefaultQuery("districtId", "")
	hasData := c.DefaultQuery("hasData", "false")

	// Tentukan kondisi berdasarkan hasData
	var hasDataCondition string
	if hasData == "true" {
		hasDataCondition = "IS NOT NULL"
	} else {
		hasDataCondition = "IS NULL"
	}

	// Query untuk mengambil data kecamatan berdasarkan kondisi
	query := db.Table("indonesia_districts").
		Select("indonesia_districts.id, indonesia_districts.name, "+
			"COUNT(DISTINCT tps.id) AS tpsCount, "+
			"SUM(CASE WHEN data_recaps.id IS NOT NULL THEN 1 ELSE 0 END) AS tpsWithData, "+
			"SUM(CASE WHEN data_recaps.id IS NULL THEN 1 ELSE 0 END) AS tpsWithoutData").
		Joins("LEFT JOIN tps ON indonesia_districts.id = tps.district_id").
		Joins("LEFT JOIN data_recaps ON tps.id = data_recaps.tps_id").
		Where("data_recaps.id "+hasDataCondition).
		Where("indonesia_districts.city_code = ?", 3202).  // Menambahkan filter untuk code_city
		Group("indonesia_districts.id")

	// Jika ada filter berdasarkan districtId
	if districtId != "" {
		districtIDInt, _ := strconv.Atoi(districtId)
		query = query.Where("indonesia_districts.id = ?", districtIDInt)
	}

	// Ambil data kecamatan
	var districts []District
	if err := query.Find(&districts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error querying the database"})
		log.Println("Error querying the database:", err)
		return
	}

	// Kirim data kecamatan ke frontend
	c.JSON(http.StatusOK, gin.H{
		"districts": districts,
	})
}
type Village struct {
    ID          int    `gorm:"column:id"`
    Name        string `gorm:"column:name"`
    TpsCount    int64  `gorm:"column:tpsCount"`
    TpsWithData int64  `gorm:"column:tpsWithData"`
    TpsWithoutData int64 `gorm:"column:tpsWithoutData"`
}


func getVillage(c *gin.Context) {
    districtCode := c.DefaultQuery("districtCode", "") // Menangkap query parameter districtCode
    hasData := c.DefaultQuery("hasData", "false")      // Menangkap query parameter hasData

    // Query untuk mengambil data desa berdasarkan districtCode
	var hasDataConditions string
   
	if hasData == "true" {
		hasDataConditions = "IS NOT NULL"
	} else {
		hasDataConditions = "IS NULL"
	}

	
	// Query untuk mengambil data kecamatan berdasarkan kondisi
	query := db.Table("indonesia_villages").
		Select("indonesia_villages.id, indonesia_villages.name, "+
			"COUNT(DISTINCT tps.id) AS tpsCount, "+
			"SUM(CASE WHEN data_recaps.id IS NOT NULL THEN 1 ELSE 0 END) AS tpsWithData, "+
			"SUM(CASE WHEN data_recaps.id IS NULL THEN 1 ELSE 0 END) AS tpsWithoutData").
		Joins("LEFT JOIN tps ON indonesia_villages.id = tps.village_id").
		Joins("LEFT JOIN data_recaps ON tps.id = data_recaps.tps_id").
		Where("data_recaps.id "+hasDataConditions).
		Where("tps.district_id = ?", districtCode). 
		Group("indonesia_villages.id")

    // Ambil data desa
    var villages []Village
    if err := query.Find(&villages).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error querying the database"})
        log.Println("Error querying the database:", err)
        return
    }

    // Kirim data desa ke frontend
    c.JSON(http.StatusOK, gin.H{
        "villages": villages,
    })
}

type TPSDataVal struct {
    TpsName string `gorm:"column:tps_name"`
    Value   bool   `gorm:"column:tps_value"` // Kolom ini menandakan apakah TPS sudah memiliki data_recaps
}
func getTpsByVillageId(c *gin.Context) {
    villageId := c.DefaultQuery("villageId", "")
    districtId := c.DefaultQuery("districtId", "")
    hasData := c.DefaultQuery("hasData", "false")

    // Tentukan kondisi berdasarkan hasData
    hasDataCondition := "IS NULL"
    if hasData == "true" {
        hasDataCondition = "IS NOT NULL"
    }

    // Query untuk mengambil data TPS berdasarkan villageId dan hasData
    query := db.Table("tps").
        Select("tps.name as tps_name, "+
            "CASE WHEN data_recaps.id IS NOT NULL THEN true ELSE false END AS tps_value").
        Joins("LEFT JOIN data_recaps ON tps.id = data_recaps.tps_id").
        Where("tps.village_id = ?", villageId).
        Where("tps.district_id = ?", districtId).
        Where("data_recaps.id "+hasDataCondition).
        Order("CAST(SUBSTRING_INDEX(tps.name, ' ', -1) AS UNSIGNED) ASC") // Mengurutkan secara numerik berdasarkan angka terakhir di nama TPS

    // Ambil data TPS
    var tpsData []TPSDataVal
    if err := query.Find(&tpsData).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error querying the database"})
        log.Println("Error querying the database:", err)
        return
    }

    // Kirim data TPS ke frontend
    c.JSON(http.StatusOK, gin.H{
        "tps": tpsData,
    })
}



func main() {
	// Inisialisasi database
	initDB()

	// Membuat router Gin
	r := gin.Default()

	// Rute untuk API
	r.GET("/api/tps", getTPSData)
	r.GET("/api/district", getDistrict)
	r.GET("/api/village", getVillage)
	r.GET("/api/village/tps", getTpsByVillageId)

	// Menjalankan server
	r.Run(":3000")
}
