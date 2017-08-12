package festival

import (
	"fmt"
	"github.com/nosixtools/solarlunar"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var RULE_PATTERN = "^\\d+_(solar|lunar)\\((?:m(\\d+)):(ld|(?:d|(?:fw|lw|w(\\d+))n)(\\d+))\\)=\\S+$"
var PATTERN = "^(solar|lunar)\\((?:m(\\d+)):(ld|(?:d|(?:fw|lw|w(\\d+))n)(\\d+))\\)$"
var MONTH_SOLAR_FESTIVAL = map[string][]string{}
var MONTH_LUNAR_FESTIVAL = map[string][]string{}
var SOLAR = "solar"
var LUNAR = "lunar"
var dateLayout = "2006-01-02"

func init() {
	for _, rule := range RULE_ARRAY {
		is, err := regexp.MatchString(RULE_PATTERN, rule)
		if err != nil {
			fmt.Println(err.Error())
		}
		if is {
			items := strings.Split(rule, "_")
			key := items[0]
			value := items[1]
			if strings.Contains(value, SOLAR) {
				if _, ok := MONTH_SOLAR_FESTIVAL[key]; ok {
					MONTH_SOLAR_FESTIVAL[key] = append(MONTH_SOLAR_FESTIVAL[key], value)
				} else {
					temp := []string{value}
					MONTH_SOLAR_FESTIVAL[key] = temp
				}
			} else {
				if _, ok := MONTH_LUNAR_FESTIVAL[key]; ok {
					MONTH_LUNAR_FESTIVAL[key] = append(MONTH_LUNAR_FESTIVAL[key], value)
				} else {
					temp := []string{value}
					MONTH_LUNAR_FESTIVAL[key] = temp
				}
			}
		}
	}
}
func GetFestivals(solarDay string) (festivals []string) {
	festivals = []string{}
	loc, _ := time.LoadLocation("Local")

	//处理公历节日
	tempDate, _ := time.ParseInLocation(dateLayout, solarDay, loc)
	for _, festival := range processRule(tempDate, MONTH_SOLAR_FESTIVAL, false, solarDay) {
		festivals = append(festivals, festival)
	}
	//处理农历节日
	lunarDate, isLeapMonth := solarlunar.SolarToLuanr(solarDay)
	if !isLeapMonth {
		tempDate, _ := time.ParseInLocation(dateLayout, lunarDate, loc)
		for _, festival := range processRule(tempDate, MONTH_LUNAR_FESTIVAL, true, solarDay) {
			festivals = append(festivals, festival)
		}
	}
	return
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
	tempDate, err := time.Parse(dateLayout, solarDay)
	if err != nil {
		fmt.Println(err.Error())
	}
	dayDuaration, _ := time.ParseDuration("24h")
	nextDate := tempDate.Add(dayDuaration)
	lunarDate, _ := solarlunar.SolarToLuanr(nextDate.Format(dateLayout))
	nexLunarDay, err := time.Parse(dateLayout, lunarDate)
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

func isDayOfLastWeeekInTheMonth(now time.Time, weekNumber int) bool {
	var endDayOfMonth time.Time
	year := now.Year()
	month := int(now.Month())
	isLeap := false
	if year%4 == 0 && year%100 != 0 || year%400 == 0 {
		isLeap = true
	}
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

var RULE_ARRAY = []string{"1_solar(m1:d1)=元旦",
	"1_solar(m1:d6)=中国13亿人口日",
	"1_solar(m1:d10)=中国110宣传日",
	"1_solar(m1:lwn1)=世界防治麻风病日",
	"2_solar(m2:d2)=世界湿地日",
	"2_solar(m2:d4)=世界抗癌症日",
	"2_solar(m2:d10)=世界气象日",
	"2_solar(m2:d14)=情人节",
	"2_solar(m2:d21)=国际母语日",
	"2_solar(m2:d29)=国际罕见病日",
	"3_solar(m3:d3)=全国爱耳日",
	"3_solar(m3:d8)=国际妇女节",
	"3_solar(m3:d12)=植树节（中国）",
	"3_solar(m3:d15)=世界消费者权益日",
	"3_solar(m3:d21)=世界森林日",
	"3_solar(m3:d22)=世界水日",
	"3_solar(m3:d23)=世界气象日",
	"3_solar(m3:d24)=世界防治结核病日",
	"4_solar(m4:d1)=愚人节",
	"4_solar(m4:d5)=清明节",
	"4_solar(m4:d7)=世界卫生日",
	"4_solar(m4:d22)=世界地球日",
	"5_solar(m5:d1)=国际劳动节",
	"5_solar(m5:d4)=中国青年节",
	"5_solar(m5:d8)=世界红十字与红新月日",
	"5_solar(m5:d12)=国际护士日",
	"5_solar(m5:d15)=国际家庭日|全国碘缺乏病宣传日",
	"5_solar(m5:d17)=世界电信和信息社会日|世界高血压日",
	"5_solar(m5:d18)=国际博物馆日",
	"5_solar(m5:d19)=中国汶川地震哀悼日",
	"5_solar(m5:d20)=全国学生营养日",
	"5_solar(m5:d22)=国际生物多样性日",
	"5_solar(m5:d31)=世界无烟日",
	"5_solar(m5:w2n1)=母亲节",
	"5_solar(m5:w3n1)=全国助残日",
	"5_solar(m5:w3n3)=国际牛奶日",
	"6_solar(m6:d1)=国际儿童节",
	"6_solar(m6:d5)=世界环境日",
	"6_solar(m6:d6)=全国爱眼日",
	"6_solar(m6:d14)=世界献血日",
	"6_solar(m6:d17)=防治荒漠化和干旱日",
	"6_solar(m6:d23)=国际奥林匹克日",
	"6_solar(m6:d25)=全国土地日",
	"6_solar(m6:d26)=国际禁毒日（反毒品日）",
	"6_solar(m6:w3n1)=父亲节",
	"7_solar(m7:d1)=建党节|香港回归纪念日",
	"7_solar(m7:d11)=世界人口日",
	"8_solar(m8:d1)=建军节",
	"8_solar(m8:d15)=抗日战争纪念日（香港）",
	"9_solar(m9:d3)=抗日战争胜利纪念日（中国大陆、台湾）",
	"9_solar(m9:d8)=国际扫盲日",
	"9_solar(m9:d10)=教师节|世界预防自杀日",
	"9_solar(m9:d16)=国际臭氧层保护日",
	"9_solar(m9:d17)=世界和平日",
	"9_solar(m9:d20)=全国爱牙日（中国大陆）",
	"9_solar(m9:d27)=世界旅游日",
	"9_solar(m9:w4n1)=国际聋人节",
	"10_solar(m10:d1)=国庆节",
	"10_solar(m10:d2)=国际减轻自然灾害日",
	"10_solar(m10:d4)=世界动物日",
	"10_solar(m10:d7)=世界住房日（世界人居日）",
	"10_solar(m10:d8)=全国高血压日|世界视觉日",
	"10_solar(m10:d9)=世界邮政日",
	"10_solar(m10:d10)=世界精神卫生日",
	"10_solar(m10:d15)=国际盲人节",
	"10_solar(m10:d16)=世界粮食节",
	"10_solar(m10:d17)=世界消除贫困日",
	"10_solar(m10:d22)=世界传统医药日",
	"10_solar(m10:d24)=联合国日",
	"10_solar(m10:d31)=万圣节",
	"11_solar(m11:d8)=记者节",
	"11_solar(m11:d9)=消防宣传日",
	"11_solar(m11:d14)=世界糖尿病日",
	"11_solar(m11:d17)=国际大学生节",
	"11_solar(m11:w4n5)=感恩节",
	"12_solar(m12:d1)=世界艾滋病日",
	"12_solar(m12:d3)=世界残疾人日",
	"12_solar(m12:d9)=世界足球日",
	"12_solar(m12:d13)=南京大屠杀死难者国家公祭日",
	"12_solar(m12:d20)=澳门回归纪念日",
	"12_solar(m12:d21)=国际篮球日",
	"12_solar(m12:d24)=平安夜",
	"12_solar(m12:d25)=圣诞节|世界强化免疫日",
	"12_solar(m12:d26)=毛泽东诞辰",
	"1_lunar(m1:d1)=春节",
	"1_lunar(m1:d5)=路神生日",
	"1_lunar(m1:d15)=元宵节",
	"2_lunar(m2:d2)=龙抬头",
	"4_lunar(m4:d4)=寒食节",
	"5_lunar(m5:d5)=端午节",
	"6_lunar(m6:d6)=天贶节|姑姑节",
	"7_lunar(m7:d7)=七夕节",
	"7_lunar(m7:d15)=中元节(鬼节)",
	"7_lunar(m7:d30)=地藏节",
	"8_lunar(m8:d15)=中秋节",
	"9_lunar(m9:d9)=重阳节",
	"10_lunar(m10:d1)=祭祖节",
	"12_lunar(m12:d8)=腊八节",
	"12_lunar(m12:d23)=小年",
	"12_lunar(m12:ld)=除夕"}
