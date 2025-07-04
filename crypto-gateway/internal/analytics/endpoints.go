package analytics

var endpoints = map[string]int{ // придумать лучший способ хранения и удалить это
	"/v3/ping":         1,
	"/v3/ticker/price": 2,
	"/v3/ticker/24hr":  80, // если с символом, можно отдельно сделать логику для 2, но сейчас это нафиг не надо
	"/v3/exchangeInfo": 20,
}
