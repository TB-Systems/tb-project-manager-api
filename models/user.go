package models

type User struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	Name     string `gorm:"not null" json:"name"`
	Username string `gorm:"uniqueIndex;not null" json:"username"`
	Password string `gorm:"not null" json:"password"`
	Email    string `gorm:"uniqueIndex;not null" json:"email"`
	CPF      string `gorm:"column:cpf;uniqueIndex;not null" json:"cpf"`
}
