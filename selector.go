package main

import "time"

type Selector interface {
	Select() *Target
}

type MinimumResponseDuration struct {
}

func (_this *MinimumResponseDuration)Select() *Target {
	var best *Target
	var bestValue time.Duration
	statusMap.Range(func(key, value interface{}) bool {
		currentValue := value.(time.Duration)
		if best == nil || bestValue > currentValue {
			best = key.(*Target)
			bestValue = currentValue
		}
		return true
	})
	return best
}
