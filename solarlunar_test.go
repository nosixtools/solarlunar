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
	lunarDate := "2020-02-30"
	fmt.Println(LunarToSolar(lunarDate, false))
}
