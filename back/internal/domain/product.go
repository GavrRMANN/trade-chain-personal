package domain

import (
	"time"
)

type CreateProductRequest struct {
	CreateProductDTO
	Wishlist *CreateWishlistDTO `json:"wishlist"`
}

type ProductStatus string

const (
	ProductActive    ProductStatus = "active"
	ProductReserved  ProductStatus = "reserved"
	ProductExchanged ProductStatus = "exchanged"
	ProductArchived  ProductStatus = "archived"
)

type Product struct {
	ProductID   string        `json:"product_id"`
	CustomerID  string        `json:"customer_id"`
	CategoryID  *string       `json:"category_id,omitempty"`
	Title       string        `json:"title"`
	Description string        `json:"description,omitempty"`
	Image       string        `json:"image,omitempty"`
	Price       int           `json:"price,omitempty"`
	Location    string        `json:"location,omitempty"`
	Status      ProductStatus `json:"status"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`

	// Matched отвечает на вопрос «почему мне это показали»: владелец товара
	// ищет что-то из того, что уже есть у смотрящего ленту, поэтому обмен
	// возможен напрямую, без цепочки. Заполняется только в ленте каталога и
	// только для авторизованного пользователя.
	Matched bool `json:"matched,omitempty"`
	// MatchedByProductID — товар смотрящего, который закрывает желание
	// владельца. Нужен, чтобы с карточки можно было сразу предложить обмен.
	MatchedByProductID *string `json:"matched_by_product_id,omitempty"`
}

type CreateProductDTO struct {
	CustomerID  string         `json:"customer_id" validate:"required"`
	CategoryID  *string        `json:"category_id"`
	Title       string         `json:"title" validate:"required"`
	Description string         `json:"description"`
	Image       string         `json:"image"`
	Price       int            `json:"price"`
	Location    string         `json:"location"`
	Status      *ProductStatus `json:"status"`
}

type UpdateProductDTO struct {
	Title       *string        `json:"title"`
	Description *string        `json:"description"`
	CategoryID  *string        `json:"category_id"`
	Image       *string        `json:"image"`
	Price       *int           `json:"price"`
	Location    *string        `json:"location"`
	Status      *ProductStatus `json:"status"`
}
