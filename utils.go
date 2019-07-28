package main

import (
	"bytes"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"unicode"

	"github.com/zgs225/alfred-youdao/alfred"
	"github.com/zgs225/youdao"
)

func toYoudaoDictUrl(q string) string {
	v := fmt.Sprintf("http://dict.youdao.com/search?q=%s&keyfrom=%s", url.QueryEscape(q), url.QueryEscape("fanyi.smartResult"))
	return v
}

func joinPhonetic(phonetic, uk, us string) string {
	buf := new(bytes.Buffer)
	empty := true
	if len(phonetic) > 0 {
		buf.WriteString(phonetic)
		empty = false
	}

	if len(uk) > 0 {
		if !empty {
			buf.WriteString("; ")
		}
		buf.WriteString("[UK] ")
		buf.WriteString(uk)
		empty = false
	}

	if len(us) > 0 {
		if !empty {
			buf.WriteString("; ")
		}
		buf.WriteString("[US] ")
		buf.WriteString(us)
	}

	return buf.String()
}

var (
	langPattern = regexp.MustCompile(`^([a-zA-Z\-]+)=>([a-zA-Z\-]+)$`)
)

// 解析传入进来的参数
// 返回值是 查询字符串，源语言，目标语言，是否设置语言
func parseArgs(args []string) (q string, from string, to string, lang bool) {
	if len(args) < 2 {
		return
	}
	if v := langPattern.FindAllStringSubmatch(strings.TrimSpace(args[1]), -1); len(v) > 0 {
		lang = true
		from = v[0][1]
		if from == "zh" {
			from = "zh-CHS"
		}
		to = v[0][2]
		if to == "zh" {
			to = "zh-CHS"
		}
		q = strings.TrimSpace(strings.Join(args[2:], " "))
	} else {
		q = strings.TrimSpace(strings.Join(args[1:], " "))
	}
	return
}

func copyModElementMap(m map[string]*alfred.ModElement) map[string]*alfred.ModElement {
	m2 := make(map[string]*alfred.ModElement)
	for k, v := range m {
		m2[k] = v
	}
	return m2
}

// FIXED: Output the voice of the original query if the to_language is Chinese.
// TODO:
// (1) In some translation, the output is mixed with eng and chinese. Perhaps we need to
// split the eng and chinese with its corresponding "-l %s %s". And if the output is ja,
// we need to bypass this split. (done!)
// (2) When from_lan is chinese, the subtitles will be rather output for 网络释义. (done!)

// func wordsToSayCmdOption(q string, r *youdao.Result) string {
func wordsToSayCmdOption(result string, query string, r *youdao.Result) string {
	ls := strings.Split(r.L, "2")
	if len(ls) >= 2 {
		l_from := languageToSayLanguage(ls[0])
		l_to := languageToSayLanguage(ls[1])
		if l_to == "zh_CN" {
			return fmt.Sprintf("-l %s %s", l_from, query)
		} else {
			if l_to != "ja_JP" {
				result_filtered := eliminateChinese(result)
				return fmt.Sprintf("-l %s %s", l_to, result_filtered)
			} else {
				return fmt.Sprintf("-l %s %s", l_to, result)
			}
		}
	}
	return result
}

func languageToSayLanguage(l string) string {
	switch l {
	case "zh-CHS":
		return "zh_CN"
	case "ja":
		return "ja_JP"
	case "en":
		return "en_US"
	case "ko":
		return "ko_KR"
	case "fr":
		return "fr_FR"
	case "ru":
		return "ru_RU"
	case "pt":
		return "pt_PT"
	case "es":
		return "es_ES"
	default:
		return l
	}
}

func eliminateChinese(result string) string {
	var result_filtered string
	result_rune := []rune(result)
	cn_char := make([]int, 0, len(result_rune))
	for i, c := range result_rune {
		// eliminate chinese and non-space symbols
		if !unicode.Is(unicode.Scripts["Han"], c) && (unicode.IsLetter(c) || unicode.IsSpace(c) || string(c) == ";") {
			cn_char = append(cn_char, i)
		}
	}
	for _, i := range cn_char {
		for k, _ := range result_rune {
			if k == i {
				result_filtered += string(result_rune[i])
			}
		}
	}
	return result_filtered
}

func containChinese(result string) bool {
	result_rune := []rune(result)
	for _, c := range result_rune {
		if unicode.Is(unicode.Scripts["Han"], c) {
			return true
		}
	}
	return false
}

func reverseSlice(s []string) []string {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}
