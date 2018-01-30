package kyuko

import (
	"errors"
	"time"

	"fmt"

	"github.com/PuerkitoBio/goquery"
	goTwitter "github.com/dghubble/go-twitter/twitter"
	"github.com/g-hyoga/kyuko/go/src/model"
	"github.com/g-hyoga/kyuko/go/src/scrape"
	"github.com/g-hyoga/kyuko/go/src/twitter"
)

func Exec(place int, client *goTwitter.Client) ([]model.KyukoData, error) {
	var kyukoData []model.KyukoData

	isTommorow := allowTommorowData()

	doc, err := readHTML(place, isTommorow)
	if err != nil {
		return kyukoData, err
	}

	kyukoData, err = scraper(doc, place)
	if err != nil {
		return kyukoData, err
	}

	var db model.DB
	err = manageDB(kyukoData, db)
	if err != nil {
		return kyukoData, err
	}

	/*
		err = manageTwitter(kyukoData, client)
		if err != nil {
			return kyukoData, err
		}
	*/

	return kyukoData, nil
}

func allowTommorowData() bool {
	//今の時間
	nowTime := time.Now().Hour()
	// 18:00超えてたら次の日の情報にする
	if nowTime >= 18 {
		return true
	}
	//今日の曜日
	weekday := int(time.Now().Weekday())
	// 日曜なら月曜の情報にする
	if weekday >= 7 {
		return true
	}
	return false
}

func weekdayToday() int {
	//今日の曜日
	weekday := int(time.Now().Weekday())
	//今の時間
	nowTime := time.Now().Hour()
	// 18:00超えてたら次の日の情報にする
	if nowTime >= 18 {
		weekday += 1
	}
	// 日曜なら月曜の情報にする
	if weekday == 7 {
		weekday = 1
	}
	return weekday
}

func readHTML(place int, isTommorow bool) (*goquery.Document, error) {
	//第一引数:校地
	//第二引数:曜日
	url, err := scrape.SetUrl(place, isTommorow)
	if err != nil {
		return nil, err
	}
	//http
	reader, err := scrape.Get(url)
	if err != nil {
		return nil, err
	}
	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		return nil, err
	}
	return doc, err
}

func scraper(doc *goquery.Document, place int) ([]model.KyukoData, error) {
	kyukoData, err := scrape.Scrape(doc, place)
	if err != nil {
		return nil, err
	}
	return kyukoData, nil
}

//Reason, Dayは一緒に扱う事が多いので
func insertReasonDay(db model.DB, id int, reason, day string) error {
	r := model.Reason{CanceledClassID: id, Reason: reason}
	_, err := db.InsertReason(r)
	if err != nil {
		return err
	}
	d := model.Day{CanceledClassID: id, Date: day}
	_, err = db.InsertDay(d)
	if err != nil {
		return err
	}
	return nil
}

func manageDB(kyukoData []model.KyukoData, db model.DB) error {
	err := db.Connect()
	if err != nil {
		return err
	}
	defer db.Close()

	for _, data := range kyukoData {
		_, err = db.Insert(data)
		if err != nil {
			return err
		}

		canceledClass, err := model.KyukoToCanceled(data)
		if err != nil {
			return err
		}

		//挿入するデータが存在するのか確認
		id, err := db.ShowCanceledClassID(canceledClass)
		if err != nil {
			return err
		}

		//DBに存在するデータで今日のデータでないなら
		if isExist, _ := db.IsExistToday(id, data.Day); id != -1 && !isExist {
			canceledClass.ID = id
			_, err = db.AddCanceled(canceledClass.ID)
			if err != nil {
				return err
			}
			//reason, dayにも追加
			err = insertReasonDay(db, id, data.Reason, data.Day)
			if err != nil {
				return err
			}

			//dbにない時
		} else if id == -1 {
			canceledClass.Canceled = 1
			_, err = db.InsertCanceledClass(canceledClass)
			if err != nil {
				return err
			}
			id, err = db.ShowCanceledClassID(canceledClass)
			if err != nil {
				return err
			}
			//reason, dayにも追加
			err = insertReasonDay(db, id, data.Reason, data.Day)
			if err != nil {
				return err
			}
		}

	}
	return nil
}

func manageTwitter(kyukoData []model.KyukoData, client *goTwitter.Client) error {
	if len(kyukoData) <= 0 {
		return errors.New("tweet content of null")
	}

	tws, err := twitter.CreateContent(kyukoData)
	if err != nil {
		return err
	}

	fmt.Println(tws)
	/*
		for _, tw := range tws {
			err := twitter.Update(client, tw)
			if err != nil {
				return err
			}
		}
	*/
	return nil
}

func kyukoToCanceled(db model.DB) error {
	k, err := db.SelectAll()
	if err != nil {
		return err
	}

	err = manageDB(k, db)
	if err != nil {
		return err
	}

	return nil
}