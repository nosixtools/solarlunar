package festival

import (
	"fmt"
	"testing"
)

func TestGetFestivals(t *testing.T) {
	fmt.Println(GetFestivals("2017-08-28"))
	fmt.Println(GetFestivals("2017-05-01"))
	fmt.Println(GetFestivals("2017-04-05"))
	fmt.Println(GetFestivals("2017-10-01"))
}
