package entity

import "time"

// Пользователь.
type User struct {
	UserID    string    `json:"user_id"`    // Идентификатор пользователя.
	CreatedAt time.Time `json:"created_at"` // Дата создания.
	Name      string    `json:"name"`       // Имя.
	Email     string    `json:"email"`      // Email. (Опционально)
	Password  string    `json:"password"`   // Пароль зашифрованный. (Опционально)
	Phone     string    `json:"phone"`      // Телефон. (Опционально)
}
