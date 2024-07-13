package entity

import "time"

type Product struct {
	ProductID string    `json:"product_id"` // Идентификатор продукта.
	Name      string    `json:"name"`       // Название продукта.
	CreatedAt time.Time `json:"created_at"` // Дата создания.
	Creator   string    `json:"creator"`    // Кто создал.
	URL       string    `json:"url"`        // URL на продукт в любом магазине.
	Proteins  float64   `json:"protein"`    // Белки.
	Fats      float64   `json:"fat"`        // Жиры.
	Carbs     float64   `json:"carbs"`      // Углеводы.
	Fibers    float64   `json:"fibers"`     // Волокна.
	Calories  float64   `json:"calories"`   // Калории в 100 г.
}
