package solarlunar

import (
	"fmt"
	"testing"
)

func TestSolarToChineseLuanr(t *testing.T) {
	solarDate := "1990-05-06"
	fmt.Println(SolarToChineseLuanr(solarDate))
}

func TestSolarToSimpleLunar(t *testing.T) {
	solarDate := "1990-05-06"
	fmt.Println(SolarToSimpleLuanr(solarDate))
}

func TestLunarToSolar(t *testing.T) {
	lunarDate := "1990-04-12"
	fmt.Println(LunarToSolar(lunarDate, false))
}
