package utils

func Unwarp[T any](result T, err error) T {
	if err != nil {
		panic(err)
	}
	return result
}

func UnwarpIgnore[T any](result T, _ any) T {
	return result
}
