package util

import (
	"fmt"
	"github.com/eddycjy/fake-useragent"
	"github.com/parnurzeal/gorequest"
	"math/rand"
	"strings"
	"time"
	"unicode/utf8"
)

func GetRandomUA() string {
	return browser.Computer()
}

// ContainAny 判断目标字符串是否在列表里面
func ContainAny(str string, list []string) bool {
	for _, n := range list {
		if strings.Contains(str, n) {
			return true
		}
	}
	return false
}

// ErrAndStatus goRequest错误判断
func ErrAndStatus(errs []error, resp gorequest.Response) (err error) {
	if len(errs) > 0 {
		err = errs[0]
		return
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("http code: %d", resp.StatusCode)
	}

	return
}

// UniqueRandomNum 生成count个[start,end)结束的不重复的随机数切片
func UniqueRandomNum(start int, end int, count int) []int {
	//范围检查
	if end < start || (end-start) < count {
		return nil
	}

	//存放结果的slice
	nums := make([]int, 0)
	//随机数生成器，加入时间戳保证每次生成的随机数不一样
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for len(nums) < count {
		//生成随机数
		num := r.Intn(end-start) + start

		//查重
		exist := false
		for _, v := range nums {
			if v == num {
				exist = true
				break
			}
		}

		if !exist {
			nums = append(nums, num)
		}
	}

	return nums
}

// TooShort Drop the short content
func TooShort(text string, min int) bool {
	if utf8.RuneCountInString(text) < min {
		return true
	}
	return false
}

// RemoveRepByLoop 通过两重循环过滤重复元素
func RemoveRepByLoop(slc []string) []string {
	var result []string // 存放结果
	for i := range slc {
		flag := true
		for j := range result {
			if slc[i] == result[j] {
				flag = false // 存在重复元素，标识为false
				break
			}
		}
		if flag { // 标识为false，不添加进结果
			result = append(result, slc[i])
		}
	}
	return result
}

func RandSecond(min, max int64) time.Duration {
	if min >= max || min == 0 || max == 0 {
		return time.Duration(max) * time.Second
	}
	return time.Duration(rand.Int63n(max-min)+min) * time.Second
}
