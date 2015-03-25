package metric

func avg(arr []float64) float64 {
	if len(arr) == 0 {
		return 0
	}

	var total float64
	for i := range arr {
		total += arr[i]
	}

	return total / float64(len(arr))
}
