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

var RULE_PATTERN = "^(solar|lunar)\\((?:m(\\d+)):(ld|(?:d|(?:fw|lw|w(\\d+))n)(\\d+))\\)=\\S+$"
var PATTERN = "^(solar|lunar)\\((?:m(\\d+)):(ld|(?:d|(?:fw|lw|w(\\d+))n)(\\d+))\\)$"
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
	month := strconv.Itoa(int(date.Month()))
	day := strconv.Itoa(date.Day())
	rules := ruleMap[month]
	for _, rule := range rules {
		items := strings.Split(rule, "=")
		reg, _ := regexp.Compile(PATTERN)
		subMatch := reg.FindStringSubmatch(items[0])
		festivalMonth := subMatch[2]
		if strings.HasPrefix(subMatch[3], "d") {
			festivalDay := subMatch[5]
			if month == festivalMonth && day == festivalDay {
				festivals = append(festivals, items[1])
			}
		} else if strings.HasPrefix(subMatch[3], "w") {
			festivalWeek := subMatch[3][1:2]
			festivalDayOfWeek := subMatch[3][3:4]
			week := strconv.Itoa(weekOfMonth(date))
			dayOfWeek := strconv.Itoa((int(date.Weekday()) + 1) % 7)
			if festivalWeek == week && festivalDayOfWeek == dayOfWeek {
				festivals = append(festivals, items[1])
			}
		} else if strings.HasPrefix(subMatch[3], "lw") {
			festivalDayOfWeek, _ := strconv.Atoi(subMatch[3][3:4])
			if isDayOfLastWeeekInTheMonth(date, festivalDayOfWeek) {
				festivals = append(festivals, items[1])
			}
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
		}
	}
	return festivals
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

func isLeapYear(year int) bool  {
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
