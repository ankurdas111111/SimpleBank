package util

import "math"

var FxRateINR = map[string]float64{
	"INR": 1.0,
	"USD": 83.0,
	"EUR": 90.0,
}

func ConvertAmount(amount int64, fromCurrency, toCurrency string) (toAmount int64, rate float64, ok bool) {
	from := FxRateINR[fromCurrency]
	to := FxRateINR[toCurrency]
	if from == 0 || to == 0 {
		return 0, 0, false
	}
	rate = from / to
	toAmount = int64(math.Round(float64(amount) * rate))
	return toAmount, rate, true
}


