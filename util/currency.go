package util

// constants for all supported currencies
const (
	INR = "INR"
)
//isSupportCurrency returns true if the currency is supported 
func IsSupportedCurrency(currency string) bool {
	switch currency {
	case USD,EUR,INR:
		return true
	}
	return false
}