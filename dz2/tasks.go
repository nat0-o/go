package homework

// PairInput соответствует объекту nums/target из условия контеста.
type PairInput struct {
	Nums   []int
	Target int
}

// longestPrefix возвращает самый длинный общий префикс строк.
func longestPrefix(words []string) string {
	if len(words) == 0 {
		return ""
	}
	if words[0] == "" {
		return ""
	}
	prefix := []rune(words[0])
	for i := 1; i < len(words); i++ {
		currentWord := []rune(words[i])
		j := 0
		for j < len(prefix) && j < len(currentWord) && prefix[j] == currentWord[j] {
			j++
		}
		prefix = prefix[:j]
		if len(prefix) == 0 {
			return ""
		}
	}
	return string(prefix)
}

// pairIndices возвращает индексы двух разных элементов по возрастанию.
func pairIndices(input PairInput) [2]int {
	for i := 0; i < len(input.Nums); i++ {
		for j := i + 1; j < len(input.Nums); j++ {
			if input.Nums[i]+input.Nums[j] == input.Target {
				return [2]int{i, j}
			}
		}
	}
	return [2]int{0, 0}
}

// hasDuplicate сообщает, встречается ли какое-либо значение повторно.
func hasDuplicate(nums []int) bool {
	seen := make(map[int]bool)
	for _, num := range nums {
		if seen[num] {
			return true
		}
		seen[num] = true
	}
	return false
}

// compactSorted создаёт независимый слайс уникальных значений.
func compactSorted(nums []int) []int {
	if len(nums) == 0 {
		return []int{}
	}
	result := make([]int, 0)
	result = append(result, nums[0])
	for i := 1; i < len(nums); i++ {
		lastIdx := len(result) - 1
		if nums[i] != result[lastIdx] {
			result = append(result, nums[i])
		}
	}
	return result
}

// balancedBrackets проверяет пары и порядок закрытия скобок.
func balancedBrackets(text string) bool {
	var stack []rune
	for _, char := range text {
		if char == '(' || char == '[' || char == '{' {
			stack = append(stack, char)
		} else {
			if len(stack) == 0 {
				return false
			}
			lastOpen := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			if char == ')' && lastOpen != '(' {
				return false
			}
			if char == ']' && lastOpen != '[' {
				return false
			}
			if char == '}' && lastOpen != '{' {
				return false
			}
		}
	}
	return len(stack) == 0
}
