package aggregator

type priceRange struct {
	MinPrice float64
	MaxPrice float64
	Percent  float64
}

var priceThresholds = []priceRange{
	{0, 1, 5},                 // монеты до $1 — 5%
	{1, 10, 3},                // от $1 до $10 — 3%
	{10, 100, 2},              // от $10 до $100 — 2%
	{100, 1000, 1},            // от $100 до $1000 — 1%
	{1000, 10000, 0.05},       // от $1000 до $10000 — 0.5%
	{10000, 50000, 0.001},     // свыше $10k — 0.1%
	{50000, 90000, 0.0001},    // свыше $50k — 0.01%
	{90000, 9999999, 0.00001}, // свыше $90k - 0.001%
}

func getPercent(price float64) float64 {
	for _, set := range priceThresholds {
		if price >= set.MinPrice && price < set.MaxPrice {
			return set.Percent
		}
	}

	return 0.0001
}
