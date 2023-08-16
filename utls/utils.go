package utls

func PanicError(err error) {
	if err != nil {
		panic(err)
	}
}

func Contains[A comparable](slice []A, target A) bool {
	for _, item := range slice {
		if item == target {
			return true
		}
	}
	return false
}
