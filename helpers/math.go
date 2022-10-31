package helpers

func NormalizeValue(min, max, value float64) float64 {
	return (value - min) / (max - min)
}
