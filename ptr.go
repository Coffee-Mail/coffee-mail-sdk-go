package coffeemail

// Ptr devolve um ponteiro para v. Útil nos campos opcionais em que o valor zero é
// significativo, como IsActive: coffeemail.Ptr(false).
func Ptr[T any](v T) *T {
	return &v
}
