# solarlunar
[![Build Status](https://api.travis-ci.org/nosixtools/solarlunar.svg?branch=master)](https://api.travis-ci.org/nosixtools/solarlunar)

阳历和阴历相互转化

## 快速开始
#### 下载和安装
	go get github.com/nosixtools/solarlunar
#### 创建 test.go
```
package main 


import (
	"github.com/nosixtools/solarlunar" 
	"fmt"
)


func main() {
	solarDate := "1990-05-06"
	fmt.Println(solarlunar.SolarToChineseLuanr(solarDate))
	fmt.Println(solarlunar.SolarToSimpleLuanr(solarDate))
	
	lunarDate := "1990-04-12"
	fmt.Println(solarlunar.LunarToSolar(lunarDate, false))
}

```

