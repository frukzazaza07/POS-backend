package models

type UserRole string

const (
	RoleAdmin   UserRole = "admin"
	RoleCashier UserRole = "cashier"
)

type User struct {
	BaseModel
	Name     string   `gorm:"not null" json:"name"`
	Email    string   `gorm:"uniqueIndex;not null" json:"email"`
	Password string   `gorm:"not null" json:"-"`
	Role     UserRole `gorm:"type:varchar(20);default:cashier;not null" json:"role"`
	IsActive bool     `gorm:"default:true" json:"is_active"`
}
