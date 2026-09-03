package entity

// User represents the users table in the database
type User struct {
	BaseEntity
	Name  string `gorm:"column:name;type:varchar(100);not null" json:"name"`
	Email string `gorm:"column:email;type:varchar(100);unique;not null" json:"email"`
}
