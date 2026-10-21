package lesson2

// FizzBuzz возвращает FizzBuzz, Fizz, Buzz или пустую строку для обычного числа.
// Для n <= 0 возвращает invalid.
func FizzBuzz(n int) string {
	// TODO: Реализовать через switch без выражения.
	switch {
	case n <= 0:
		return "invalid"
	case (n > 0 && n%15 == 0):
		return "FizzBuzz"
	case n%3 == 0:
		return "Fizz"
	case n%5 == 0:
		return "Buzz"
	case n > 0:
		return ""
	default:
		return ""
	}
}

// Sum возвращает сумму. Для пустого среза и nil возвращает 0.
func Sum(values []int64) int64 {
	// TODO: Пройти по значениям циклом.
	if len(values) == 0 {
		return 0
	}
	var summ int64
	for _, val := range values {
		summ += val
	}
	return summ
}

// Min возвращает минимум. Для пустого среза и nil возвращает 0 по условию задачи.
func Min(values []int64) int64 {
	// TODO: Учесть пустой ввод, затем сравнить элементы.
	if len(values) == 0 {
		return 0
	}
	var minimum int64
	minimum = values[0]
	for _, val := range values {
		if val < minimum {
			minimum = val
		}
	}
	return minimum
}

// Max возвращает максимум. Для пустого среза и nil возвращает 0 по условию задачи.
func Max(values []int64) int64 {
	// TODO: Учесть пустой ввод, затем сравнить элементы.
	if len(values) == 0 {
		return 0
	}
	var maximum int64
	maximum = values[0]
	for _, val := range values {
		if val > maximum {
			maximum = val
		}
	}
	return maximum
}

// Average возвращает среднее как float64. Для пустого среза и nil возвращает 0.
func Average(values []int64) float64 {
	// TODO: Выполнить преобразования до деления.
	if len(values) == 0 {
		return 0
	}
	var average float64
	average = 0.0
	var count float64
	count = 0.0
	for _, val := range values {
		average += float64(val)
		count++
	}

	return average / count
}

// Category возвращает invalid, small, regular или large по сумме в копейках.
func Category(amount int64) string {
	// TODO: Проверить границы через switch без выражения.
	switch {
	case amount < 0:
		return "invalid"
	case amount >= 0 && amount <= 9999:
		return "small"
	case amount >= 10000 && amount <= 49999:
		return "regular"
	default:
		return "large"
	}
}

// RuneCount считает кодовые точки Unicode в строке через range.
func RuneCount(text string) int {
	// TODO: Не использовать len(text) как количество кодовых точек.
	count := 0
	for range text {
		count++
	}
	return count
}
