package sqlite

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"strconv"
	"strings"
	"time"
	"weixinScraperSingle/model"
	"weixinScraperSingle/util"
)

var (
	logger = util.GetLogger("sqlite")
)

func DbInit() {
	util.DB = *openDB("db.db")
}

func openDB(dbPath string) *sql.DB {
	// 判断数据库是否存在
	_, err := os.Stat(dbPath) //os.Stat获取文件信息
	if err != nil {
		panic(err)
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		panic(err)
	}

	return db
}

func Add(word string) bool {
	stmt, err := util.DB.Prepare("INSERT INTO sina(word) values(?)")
	if err != nil {
		logger.Debugf("添加关键词 %s 时出现错误 %s", word, err)
		return false
	}
	defer stmt.Close()

	result, err := stmt.Exec(word)

	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			return false
		}

		logger.Debugf("添加关键词 %s 时出现错误 %s", word, err)
		return false
	}

	if result != nil {
		return true
	}
	return false
}

func AddMp(mp model.WeixinMp) bool {
	stmt, err := util.DB.Prepare("INSERT INTO author(mp, mp_name) values(?, ?)")
	if err != nil {
		logger.Debugf("添加公众号 %s 时出现错误 %s", mp.Mp, err)
		return false
	}
	defer stmt.Close()

	result, err := stmt.Exec(mp.Mp, mp.MpName)

	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			return false
		}

		logger.Debugf("添加公众号 %s 时出现错误 %s", mp.Mp, err)
		return false
	}

	if result != nil {
		return true
	}
	return false
}

// 检查新浪提取的关键词
func SelectMpByName(mpName string) (exist bool) {
	sqlStr := fmt.Sprintf("SELECT id FROM 'author' WHERE mp_name = '%s'", mpName)
	rows, err := util.DB.Query(sqlStr)
	if err != nil {
		logger.Debugf("查询公众号时出现错误 %s", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		err = rows.Scan(&id)
		if id > 0 {
			return true
		}
	}

	return
}

func SelectMps() (mps []model.WeixinMp) {
	m, _ := time.ParseDuration("-10m") // 跳过10分钟内检查过的
	conditionTime := strconv.FormatInt(time.Now().Add(m).Unix(), 10)

	rows, err := util.DB.Query("SELECT mp, mp_name, renew_time FROM 'author' WHERE check_time < " + conditionTime)
	if err != nil {
		logger.Debugf("查询公众号时出现错误 %s", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var mp model.WeixinMp
		err = rows.Scan(&mp.Mp, &mp.MpName, &mp.RenewTime)
		if err != nil {
			continue
		}
		mps = append(mps, mp)
	}

	return
}

// 更新最新文章发布时间
func UpdatePostTime(mp string, timestamp string) bool {
	stmt, err := util.DB.Prepare("UPDATE 'author' set renew_time=? where mp=?")
	if err != nil {
		logger.Debugf("更新 renew_time 出现错误 %s", err)
		return false
	}
	defer stmt.Close()

	result, err := stmt.Exec(timestamp, mp)

	if err != nil {
		logger.Debugf("更新 renew_time 出现错误 %s", err)
		return false
	}

	if result != nil {
		return true
	}
	return false
}

// 更新公众号检查时间
func UpdateCheckTime(mp string, timestamp int64) bool {
	stmt, err := util.DB.Prepare("UPDATE 'author' set check_time=? where mp=?")
	if err != nil {
		logger.Debugf("更新 check_time 出现错误 %s", err)
		return false
	}
	defer stmt.Close()

	result, err := stmt.Exec(timestamp, mp)

	if err != nil {
		logger.Debugf("更新 check_time 出现错误 %s", err)
		return false
	}

	if result != nil {
		return true
	}
	return false
}

// 更新关键词检查时间
func UpdateWordCheckTime(word string, timestamp int64) bool {
	stmt, err := util.DB.Prepare("UPDATE 'sina' set check_time=? where word=?")
	if err != nil {
		logger.Debugf("更新 check_time 出现错误 %s", err)
		return false
	}
	defer stmt.Close()

	result, err := stmt.Exec(timestamp, word)

	if err != nil {
		logger.Debugf("更新关键词 check_time 出现错误 %s", err)
		return false
	}

	if result != nil {
		return true
	}
	return false
}

// 检查新浪提取的关键词
func SelectWord(word string) {
	rows, err := util.DB.Query("SELECT * FROM 'sina' WHERE word = " + word)
	if err != nil {
		logger.Debugf("查询关键词时出现错误 %s", err)
		return
	}
	defer rows.Close()

	var (
		id        int
		w         string
		checkTime int64
	)
	err = rows.Scan(&id, &w, &checkTime)

	return
}
