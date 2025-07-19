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

func (_this *MinimumResponseDuration) Select() *Target {
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

type MinCost struct {
}

func (_this *MinCost) Select() *Target {
	var best *Target
	var bestValue time.Duration
	var available []string
	statusMap.Range(func(key, value interface{}) bool {
		targetStatus := value.(TargetStatus)
		target := targetStatus.Target
		available = append(available, target.Server)
		elapsed := targetStatus.Elapsed
		var extraCost time.Duration
		if target.Params != nil {
			if extraCostText, ok := target.Params["extra-cost"]; ok {
				var err error
				extraCost, err = time.ParseDuration(extraCostText)
				if err == nil {
					elapsed += extraCost
				}
			}
		}
		if best == nil || bestValue > elapsed {
			best = target
			bestValue = elapsed
		}
		return true
	})
	bestServer := "none"
	if best != nil {
		bestServer = best.Server
	}
	_this.display(available, bestServer)
	return best
}

func (_this *MinCost) display(availableServers []string, bestServer string) {
	var display []string
	for _, s := range availableServers {
		if s == bestServer {
			display = append(display, "\033[32m"+s+"\033[0m") // 绿色
		} else {
			display = append(display, s)
		}
	}
	log.Printf("available: [%s]", strings.Join(display, ", "))
}
