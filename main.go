package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/zgs225/alfred-youdao/alfred"
	"github.com/zgs225/youdao"
)

const (
	APPID     = ""
	APPSECRET = ""
	MAX_LEN   = 255

	UPDATECMD = "alfred-youdao:update"
)

func init() {
	log.SetPrefix("[i] ")
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {
	log.Println(os.Args)

	items := alfred.NewResult()

	appID := os.Getenv("zhiyun_id")
	appKey := os.Getenv("zhiyun_key")

	if appID == "" || appKey == "" {
		items.Append(&alfred.ResultElement{
			Valid:    false,
			Title:    "错误: 请设置有道API",
			Subtitle: "有道词典",
		})
		items.End()
	}

	client := &youdao.Client{
		AppID:     appID,
		AppSecret: appKey,
	}
	agent := newAgent(client)
	q, from, to, lang := parseArgs(os.Args)

	if lang {
		if err := agent.Client.SetFrom(from); err != nil {
			items.Append(&alfred.ResultElement{
				Valid:    true,
				Title:    fmt.Sprintf("错误: 源语言不支持[%s]", from),
				Subtitle: `有道词典`,
			})
			items.End()
		}
		if err := agent.Client.SetTo(to); err != nil {
			items.Append(&alfred.ResultElement{
				Valid:    true,
				Title:    fmt.Sprintf("错误: 目标语言不支持[%s]", to),
				Subtitle: `有道词典`,
			})
			items.End()
		}
	}

	if len(q) == 0 {
		items.Append(&alfred.ResultElement{
			Valid:    true,
			Title:    "有道词典",
			Subtitle: `查看"..."的解释或翻译`,
		})
		items.End()
	}

	if len(q) > 255 {
		items.Append(&alfred.ResultElement{
			Valid:    false,
			Title:    "错误: 最大查询字符数为255",
			Subtitle: q,
		})
		items.End()
	}

	r, err := agent.Query(q)
	if err != nil {
		panic(err)
	}

	mod := map[string]*alfred.ModElement{
		alfred.Mods_Shift: &alfred.ModElement{
			Valid:    true,
			Arg:      toYoudaoDictUrl(q),
			Subtitle: "回车键打开词典网页",
		},
	}
	if r.Basic != nil {
		phonetic := joinPhonetic(r.Basic.Phonetic, r.Basic.UkPhonetic, r.Basic.UsPhonetic)
		for _, title := range r.Basic.Explains {
			mod2 := copyModElementMap(mod)
			mod2[alfred.Mods_Cmd] = &alfred.ModElement{
				Valid:    true,
				Arg:      wordsToSayCmdOption(title, q, r),
				Subtitle: "发音",
			}
			item := alfred.ResultElement{
				Valid:    true,
				Title:    title,
				Subtitle: phonetic,
				Arg:      title,
				Mods:     mod2,
			}
			items.Append(&item)

		}
	}

	if r.Translation != nil {
		title := strings.Join(*r.Translation, "; ")
		mod2 := copyModElementMap(mod)
		mod2[alfred.Mods_Cmd] = &alfred.ModElement{
			Valid:    true,
			Arg:      wordsToSayCmdOption(title, q, r),
			Subtitle: "发音",
		}
		item := alfred.ResultElement{
			Valid:    true,
			Title:    title,
			Subtitle: "翻译结果",
			Arg:      title,
			Mods:     mod2,
		}
		items.Append(&item)
	}

	if r.Web != nil {
		items.Append(&alfred.ResultElement{
			Valid:    true,
			Title:    "网络释义",
			Subtitle: "有道词典 for Alfred",
		})

		for _, elem := range *r.Web {
			mod2 := copyModElementMap(mod)
			mod2[alfred.Mods_Cmd] = &alfred.ModElement{
				Valid: true,
				// Arg:      wordsToSayCmdOption(elem.Key, q, r),
				Arg:      wordsToSayCmdOption(strings.Join(elem.Value, "; "), q, r),
				Subtitle: "发音",
			}
			ls := strings.Split(r.L, "2")
			l_from := languageToSayLanguage(ls[0])
			if l_from == "zh_CN" {
				items.Append(&alfred.ResultElement{
					Valid:    true,
					Title:    elem.Key,
					Subtitle: strings.Join(elem.Value, "; "),
					Arg:      elem.Key,
					Mods:     mod2,
					// QuickLookUrl: toYoudaoDictUrl(q),
				})
			} else {
				items.Append(&alfred.ResultElement{
					Valid:    true,
					Title:    elem.Key,
					Subtitle: strings.Join(elem.Value, "; "),
					Arg:      elem.Key,
					Mods:     mod,
					// QuickLookUrl: toYoudaoDictUrl(q),
				})

			}
		}
	}

	if agent.Dirty {
		if err := agent.Cache.SaveFile(CACHE_FILE); err != nil {
			log.Println(err)
		}
	}

	items.End()
}
