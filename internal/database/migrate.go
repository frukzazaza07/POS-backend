package database

import (
	"log"
	"os"
	"pos-backend/internal/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.User{},
		&models.POSProduct{},
		&models.Order{},
		&models.OrderItem{},
		&models.StockCache{},
		&models.BankQRConfig{},
		&models.VatConfig{},
	)
}

func Seed(db *gorm.DB) {
	if os.Getenv("SEED") != "true" {
		return
	}

	var count int64
	db.Model(&models.User{}).Count(&count)
	if count > 0 {
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("seed: failed to hash password: %v", err)
		return
	}

	admin := models.User{
		Name:     "Admin",
		Email:    "admin@pos.local",
		Password: string(hashed),
		Role:     models.RoleAdmin,
		IsActive: true,
	}
	if err := db.Create(&admin).Error; err != nil {
		log.Printf("seed: failed to create admin: %v", err)
		return
	}
	log.Println("seed: admin@pos.local / admin123")
}
