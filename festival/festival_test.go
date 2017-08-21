package festival

import (
	"fmt"
	"testing"
)

func TestGetFestivals(t *testing.T) {
	festival := NewFestival("./festival.json")
	fmt.Println(festival.GetFestivals("2017-08-28"))
	fmt.Println(festival.GetFestivals("2017-05-01"))
	fmt.Println(festival.GetFestivals("1902-04-06"))
	fmt.Println(festival.GetFestivals("2017-04-04"))
	fmt.Println(festival.GetFestivals("2017-04-03"))
	fmt.Println(festival.GetFestivals("2018-02-15"))
	fmt.Println(festival.GetFestivals("2018-02-16"))
}
