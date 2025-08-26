package errors

import "errors"

var (
	ErrProductNotFound     = errors.New("product not found")
	ErrInvalidCouponCode   = errors.New("invalid coupon code")
	ErrInvalidProductID    = errors.New("invalid product ID")
	ErrInternalServerError = errors.New("internal server error")
	ErrDatabaseError       = errors.New("database query returned error")
)

func New(s string) error {
	return errors.New(s)
}

func Is(err error, err2 error) bool {
	return errors.Is(err, err2)
}

func Join(err ...error) error {
	return errors.Join(err...)
}
