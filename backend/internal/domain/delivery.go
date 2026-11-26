package domain

func DeliveryFee(method string, quantity int) (float64, bool) {
	if quantity < 1 {
		quantity = 1
	}
	switch method {
	case "pickup":
		return 0, true
	case "freight":
		return float64(2800+45*quantity) / 100, true
	case "express":
		return float64(6400+90*quantity) / 100, true
	default:
		return 0, false
	}
}
