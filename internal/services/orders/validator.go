package orders

import "strconv"

type OrderValidator interface {
	OrderNumberIsValid(string) bool
}

type luhnValidator struct {
}

func (v luhnValidator) OrderNumberIsValid(number string) bool {
	id, err := strconv.Atoi(number)
	if err != nil {
		return false
	}

	return (id%10+v.checkSum(id/10))%10 == 0
}

func (v luhnValidator) checkSum(number int) int {
	var luhn int

	for i := 0; number > 0; i++ {
		cur := number % 10

		if i%2 == 0 {
			cur = cur * 2
			if cur > 9 {
				cur = cur%10 + cur/10
			}
		}

		luhn += cur
		number = number / 10
	}

	sum := luhn % 10
	if sum == 0 {
		return 0
	}
	return luhn % 10
}
