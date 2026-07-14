package models

type Movie struct {
	ID          uint
	Title       string `gorm:"size:100;unique"`
	Genre       string
	ReleasedAt  string
	Description string
	Rating      float64 `gorm:"type:numeric(3,1)"`
}
