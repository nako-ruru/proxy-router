package main

import (
	"log"
	"strings"
	"time"
)

type Selector interface {
	Select() *Target
}

type MinimumResponseDuration struct {
}

func (_this *MinimumResponseDuration)Select() *Target {
	var best *Target
	var bestValue time.Duration
	var available []string
	statusMap.Range(func(key, value interface{}) bool {
		target := key.(Target)
		available = append(available, target.Server)
		currentValue := value.(time.Duration)
		if best == nil || bestValue > currentValue {
			best = &target
			bestValue = currentValue
		}
		return true
	})
	bestServer := "none"
	if best != nil {
		bestServer = best.Server
	}
	log.Printf("available: [%s], select %s", strings.Join(available, ", "), bestServer)
	return best
}
