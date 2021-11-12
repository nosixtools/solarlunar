package festival

import (
	"fmt"
	"github.com/bitly/go-simplejson"
	"github.com/nosixtools/solarlunar"
	"io/ioutil"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var RULE_PATTERN = "^(solar|lunar)\\((?:m(\\d+)):(ld|(?:d|(?:fw|lw|w(\\d+))n|(?:s\\d+))(\\d+))\\)=\\S+$"
var PATTERN = "^(solar|lunar)\\((?:m(\\d+)):(ld|(?:d|(?:fw|lw|w(\\d+))n|(?:s\\d+))(\\d+))\\)$"
var MONTH_SOLAR_FESTIVAL = map[string][]string{}
var MONTH_LUNAR_FESTIVAL = map[string][]string{}
var SOLAR = "solar"
var LUNAR = "lunar"
var DATELAYOUT = "2006-01-02"

type Festival struct {
	filename string
}

func NewFestival(filename string) *Festival {
	if filename == "" {
		filename = "./festival.json"
	}
	readFestivalRuleFromFile(filename)
	return &Festival{filename: filename}
}

func (f *Festival) GetFestivals(solarDay string) (festivals []string) {
	festivals = []string{}
	loc, _ := time.LoadLocation("Local")

	//处理公历节日
	tempDate, _ := time.ParseInLocation(DATELAYOUT, solarDay, loc)
	for _, festival := range processRule(tempDate, MONTH_SOLAR_FESTIVAL, false, solarDay) {
		festivals = append(festivals, festival)
	}
	//处理农历节日
	lunarDate, isLeapMonth := solarlunar.SolarToLuanr(solarDay)
	if !isLeapMonth {
		tempDate, _ := time.ParseInLocation(DATELAYOUT, lunarDate, loc)
		for _, festival := range processRule(tempDate, MONTH_LUNAR_FESTIVAL, true, solarDay) {
			festivals = append(festivals, festival)
		}
	}
	return
}

func readFestivalRuleFromFile(filename string) {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	rules, err := simplejson.NewJson(bytes)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	solarData := rules.Get(SOLAR)
	if solarData != nil {
		solarMap, err := solarData.Map()
		if err != nil {
			fmt.Println(err.Error())
		}
		for key, value := range solarMap {
			for _, item := range value.([]interface{}) {
				v := item.(string)
				is, err := regexp.MatchString(RULE_PATTERN, v)
				if err != nil {
					fmt.Println(err.Error())
				}
				if is {
					if _, ok := MONTH_SOLAR_FESTIVAL[key]; ok {
						MONTH_SOLAR_FESTIVAL[key] = append(MONTH_SOLAR_FESTIVAL[key], v)
					} else {
						temp := []string{v}
						MONTH_SOLAR_FESTIVAL[key] = temp
					}
				}
			}
		}
	}
	lunarData := rules.Get(LUNAR)
	if lunarData != nil {
		lunarMap, err := lunarData.Map()
		if err != nil {
			fmt.Println(err.Error())
		}
		for key, value := range lunarMap {
			for _, item := range value.([]interface{}) {
				v := item.(string)
				is, err := regexp.MatchString(RULE_PATTERN, v)
				if err != nil {
					fmt.Println(err.Error())
				}
				if is {
					if _, ok := MONTH_LUNAR_FESTIVAL[key]; ok {
						MONTH_LUNAR_FESTIVAL[key] = append(MONTH_LUNAR_FESTIVAL[key], v)
					} else {
						temp := []string{v}
						MONTH_LUNAR_FESTIVAL[key] = temp
					}
				}
			}
		}
	}
}

func processRule(date time.Time, ruleMap map[string][]string, isLunar bool, solarDay string) []string {
	festivals := []string{}
	year := int(date.Year())
	month := strconv.Itoa(int(date.Month()))
	day := strconv.Itoa(date.Day())
	rules := ruleMap[month]
	for _, rule := range rules {
		items := strings.Split(rule, "=")
		reg, _ := regexp.Compile(PATTERN)
		subMatch := reg.FindStringSubmatch(items[0])
		festivalMonth := subMatch[2]
		if strings.HasPrefix(subMatch[3], "s456") && !isLunar { //特殊处理清明节
			festivalDay := getQingMingFestival(year)
			if month == festivalMonth && day == festivalDay {
				festivals = append(festivals, items[1])
			}
			continue
		} else if strings.HasPrefix(subMatch[3], "s345") && !isLunar { //特殊处理寒食节，为清明节前一天
			festivalDay := getQingMingFestival(year)
			intValue, err := strconv.Atoi(festivalDay)
			if err != nil {
				fmt.Print(err.Error())
				continue
			}
			festivalDay = strconv.Itoa(intValue - 1)
			if month == festivalMonth && day == festivalDay {
				festivals = append(festivals, items[1])
			}
		} else if strings.HasPrefix(subMatch[3], "d") {
			festivalDay := subMatch[5]
			if month == festivalMonth && day == festivalDay {
				festivals = append(festivals, items[1])
			}
			continue
		} else if strings.HasPrefix(subMatch[3], "w") {
			festivalWeek, _ := strconv.Atoi(subMatch[3][1:2])
			festivalDayOfWeek, _ := strconv.Atoi(subMatch[3][3:4])
			week := 0
			tempDayOfWeek := getDayOfWeekOnFirstDayOfMonth(date)
			//特殊处理感恩节，感恩节（m11:w4n5）的计算，不是第4周周4，而是第4个周四，如果第一个周没有周四，就不算第一周
			if (compareWeek(tempDayOfWeek, festivalDayOfWeek) && strings.HasPrefix(subMatch[3], "w4n5")) {
				week = weekOfMonth(date) - 1
			} else {
				week = weekOfMonth(date)
			}
			dayOfWeek := (int(date.Weekday()) + 1) % 7
			if festivalWeek == week && festivalDayOfWeek == dayOfWeek {
				festivals = append(festivals, items[1])
			}
			continue
		} else if strings.HasPrefix(subMatch[3], "lw") {
			festivalDayOfWeek, _ := strconv.Atoi(subMatch[3][3:4])
			if isDayOfLastWeeekInTheMonth(date, festivalDayOfWeek) {
				festivals = append(festivals, items[1])
			}
			continue
		} else if strings.HasPrefix(subMatch[3], "ld") && isLunar { //特殊处理除夕节日
			if month == "12" && day == "29" {
				nextLunarDay := lunarDateAddOneDay(solarDay)
				newMonth := strconv.Itoa(int(nextLunarDay.Month()))
				if month != newMonth {
					festivals = append(festivals, items[1])
				}
			} else if month == "12" && day == "30" {
				festivals = append(festivals, items[1])
			}
			continue
		}
	}
	return festivals
}
// 清明节算法 公式：int((yy*d+c)-(yy/4.0)) 公式解读：y=年数后2位，d=0.2422，1=闰年数，21世纪c=4081，20世纪c=5.59
func getQingMingFestival(year int) string {
	var val float64
	if year >= 2000 { //21世纪
		val = 4.81
	} else { //20世纪
		val = 5.59
	}
	d := float64(year % 100)
	day := int(d*0.2422 + val - float64(int(d)/4))
	return strconv.Itoa(day)
}

func lunarDateAddOneDay(solarDay string) time.Time {
	tempDate, err := time.Parse(DATELAYOUT, solarDay)
	if err != nil {
		fmt.Println(err.Error())
	}
	dayDuaration, _ := time.ParseDuration("24h")
	nextDate := tempDate.Add(dayDuaration)
	lunarDate, _ := solarlunar.SolarToLuanr(nextDate.Format(DATELAYOUT))
	nexLunarDay, err := time.Parse(DATELAYOUT, lunarDate)
	if err != nil {
		fmt.Println(err.Error())
	}
	return nexLunarDay
}

func weekOfMonth(now time.Time) int {
	beginningOfTheMonth := time.Date(now.Year(), now.Month(), 1, 1, 1, 1, 1, time.UTC)
	_, thisWeek := now.ISOWeek()
	_, beginningWeek := beginningOfTheMonth.ISOWeek()
	return 1 + thisWeek - beginningWeek
}

func isLeapYear(year int) bool {
	if year%4 == 0 && year%100 != 0 || year%400 == 0 {
		return true
	}
	return false
}

func isDayOfLastWeeekInTheMonth(now time.Time, weekNumber int) bool {
	var endDayOfMonth time.Time
	year := now.Year()
	month := int(now.Month())
	isLeap := isLeapYear(year)
	if month == 2 {
		if isLeap {
			endDayOfMonth = time.Date(now.Year(), now.Month(), 29, 23, 59, 59, 1, time.UTC)
		} else {
			endDayOfMonth = time.Date(now.Year(), now.Month(), 28, 23, 59, 59, 1, time.UTC)
		}
	} else if month == 1 || month == 3 || month == 5 || month == 7 || month == 8 || month == 10 || month == 12 {
		endDayOfMonth = time.Date(now.Year(), now.Month(), 31, 23, 59, 59, 1, time.UTC)
	} else {
		endDayOfMonth = time.Date(now.Year(), now.Month(), 30, 23, 59, 59, 1, time.UTC)
	}
	_, lastWeekOfMonth := endDayOfMonth.ISOWeek()
	_, nowWeekOfMonth := now.ISOWeek()
	dayOfWeek := (int(endDayOfMonth.Weekday()) + 1) % 7
	if dayOfWeek > weekNumber && lastWeekOfMonth > nowWeekOfMonth {
		dayDuaration, _ := time.ParseDuration("-24h")
		endDayOfMonth = endDayOfMonth.Add(dayDuaration * time.Duration(7))
		_, lastWeekOfMonth = endDayOfMonth.ISOWeek()
	}
	if lastWeekOfMonth == nowWeekOfMonth {
		nowDayOfWeek := (int(now.Weekday()) + 1) % 7
		if nowDayOfWeek == weekNumber {
			return true
		}
	}
	return false
}


func compareWeek(first int, second int) bool {
	if first-1 == 0 {
		first = 7;
	} else {
		first = first - 1
	}
	if second-1 == 0 {
		second = 7
	} else {
		second = second - 1
	}

	if (first >= second) {
		return true
	} else {
		return false
	}
}

// 星期日：1 星期一：2 类推
func getDayOfWeekOnFirstDayOfMonth(date time.Time) int {
	date = getFirstDateOfMonth(date)
	dayOfWeek := (int(date.Weekday()) + 1) % 7
	return dayOfWeek
}

func getFirstDateOfMonth(d time.Time) time.Time {
	tempDate := time.Date(d.Year(), d.Month(), d.Day(), d.Hour(), d.Minute(), d.Second(), d.Nanosecond(), d.Location())
	d = tempDate.AddDate(0, 0, -d.Day()+1)
	return getZeroTime(d)
}

func getLastDateOfMonth(d time.Time) time.Time {
	return getFirstDateOfMonth(d).AddDate(0, 1, -1)
}

func getZeroTime(d time.Time) time.Time {
	return time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, d.Location())
}
