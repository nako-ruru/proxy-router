package main

import "time"

type Selector interface {
	Select() *BackendC
}

type MinimumResponseDuration struct {
}

func (_this *MinimumResponseDuration)Select() *BackendC {
	var best *BackendC
	var bestValue time.Duration
	statusMap.Range(func(key, value interface{}) bool {
		currentValue := value.(time.Duration)
		if best == nil || bestValue > currentValue {
			best = key.(*BackendC)
			bestValue = currentValue
		}
		return true
	})
	return best
}
