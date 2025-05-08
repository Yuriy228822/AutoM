package models

import "errors"

// Пример функции валидации цены
func ValidatePrice(price float64) error {
	if price <= 0 {
		return errors.New("цена должна быть больше 0")
	}
	return nil
}
