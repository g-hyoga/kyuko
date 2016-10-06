package scrape

import (
	"errors"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	iconv "github.com/djimenez/iconv-go"
	"github.com/g-hyoga/kyuko/go/model"
)

//place(1: 今出川 ,2: 京田辺), week(1 ~ 6: Mon ~ Sat)を引数に持ち
//urlを生成する
func SetUrl(place, week int) (string, error) {
	url := "http://duet.doshisha.ac.jp/info/KK1000.jsp?katei=1"
	//weekに7(Sunday)はない
	if (place != 1 && place != 2) || week < 1 || week > 6 {
		return "", errors.New("place is 1 or 2, 0 < week < 7")
	} else {
		url = url + "&youbi=" + strconv.Itoa(week)
		url = url + "&kouchi=" + strconv.Itoa(place)
		return url, nil
	}
}

//校地と曜日の情報を含んだurlを引数としてとり、休講structのsliceを返す
//urlはstaticなfileを指定しても良い(test用)
func Scrape(url string) ([]model.KyukoData, error) {
	var kyukoData []model.KyukoData
	var err error

	doc, err := goquery.NewDocument(url)
	if err != nil {
		return kyukoData, err
	}

	//ここのエラー処理どうしよう
	doc.Find("tr.style1").Each(func(i int, s *goquery.Selection) {
		var k model.KyukoData

		//"1講時"みたいなのが取れる
		rawPeriod := s.Find("th.style2").Text()
		rawPeriod, err = iconv.ConvertString(rawPeriod, "shift-jis", "utf-8")
		rawPeriod = strings.Split(rawPeriod, "講時")[0]
		k.Period, err = strconv.Atoi(rawPeriod)
		if k.Period == 0 {
			k.Period = kyukoData[i-1].Period
		}

		//classがないのでこうするしかない
		tds := s.Find("td")
		for i := range tds.Nodes {
			tmp, _ := iconv.ConvertString(tds.Eq(i).Text(), "shift-jis", "utf-8")
			//授業名の時
			if i%3 == 0 {
				k.ClassName = tmp
				//講師の時
			} else if i%3 == 1 {
				//TrimSpaceとかじゃきかない
				k.Instructor = strings.Split(tmp, " ")[0] + strings.Split(tmp, " ")[4]
			}
		}

		//休講理由
		rawReason := strings.Split(s.Find("td.style3").Text(), "&")[0]
		reason, _ := iconv.ConvertString(rawReason, "shift-jis", "utf-8")
		k.Reason = strings.Split(reason, "ﾂ")[0]

		rawPlaceDayWeek := doc.Find("tr.styleT > th").Text()
		//Place
		rawPlace := strings.Split(strings.Split(rawPlaceDayWeek, "[")[1], "]")[0]
		place, _ := iconv.ConvertString(rawPlace, "shift-jis", "utf-8")
		if place == "今出川" {
			k.Place = 1
		} else if place == "京田辺" {
			k.Place = 2
		}

		//日付と曜日取らないといけない

		kyukoData = append(kyukoData, k)
	})

	return kyukoData, err
}
