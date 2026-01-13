package product

import "errors"

var (
	ErrProductNotFound  = errors.New("product not found")
	ErrInvalidProduct   = errors.New("invalid stafproductf")
	ErrDuplicateProduct = errors.New("duplicate product")
)
