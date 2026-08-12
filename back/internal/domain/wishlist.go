package domain

import (
	"time"
)

type CreateWishlistDTO struct {
	Name        string   `json:"name"`
	CategoryIDs []string `json:"category_ids"`
}
type Wishlist struct {
	WishlistID string    `json:"wishlist_id"`
	ProductID  string    `json:"product_id"`
	Name       string    `json:"name"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type WishlistOption struct {
	WishlistID string `json:"wishlist_id"`
	CategoryID string `json:"category_id"`
}
