package wash

import (
	"regexp"
	"strings"
	"weixinScraperSingle/util"
)

var (
	logger = util.GetLogger("wash")
)

func Wash(content string) string {
	content = trimHtml(content) // 清洗html

	phasesList := combination(content) // 组合段落

	phasesList = dropRepeat(&phasesList) // 丢弃重复的段落

	return strings.Join(phasesList, "\r\n")
}

// 组合段落
func combination(content string) []string {
	phaseSlice := strings.Split(content, "\n")

	var newPhases []string
	for _, p := range phaseSlice {
		p = strings.TrimSpace(p)

		// 出现下列词, 后面的段落全部丢弃
		if util.ContainAny(p, util.EndHere) {
			break
		}

		// 丢弃无用段
		if util.ContainAny(p, util.ArticleFilter) || util.TooShort(p, 2) {
			continue
		}

		//if commaRepeat(p) {
		//	continue
		//}

		newPhases = append(newPhases, p)
	}

	return newPhases

	//return strings.Join(newPhases, "\r")
}

func trimHtml(src string) string {
	//去除所有尖括号内的HTML代码，并换成换行符
	re, _ := regexp.Compile("\\<[\\S\\s]+?\\>")
	src = re.ReplaceAllString(src, "\n")

	//去除连续的换行符
	re, _ = regexp.Compile("\\s{2,}")
	src = re.ReplaceAllString(src, "\n")
	return strings.TrimSpace(src)
}

// 清理重复的段落
func dropRepeat(phases *[]string) (newPhases []string) {
	for k, phase := range *phases {
		keepIt := true

		for j, p := range *phases {
			if k == j {
				// 跳过对自身的检查
				continue
			}

			if strings.Contains(phase, p) {
				keepIt = false
			}
		}

		if keepIt {
			// 通过循环遍历每个段落, 丢弃重复的段落
			newPhases = append(newPhases, phase)
		}
	}
	return newPhases
}

// 符号重复, 字符重复的段落
func commaRepeat(phase string) bool {
	re := regexp.MustCompile(`^(.)\1+$`)

	if re.MatchString(phase) {
		return true
	}
	return false
}
