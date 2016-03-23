package main

import (
	"encoding/json"
	"fmt"
	"io"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/reusee/vviccommon"

	"net/http"
	"os"
)

var (
	pt = fmt.Printf
)

func main() {
	pagePath := fmt.Sprintf("http://www.vvic.com/api/item/%s", os.Args[1])
	resp, err := http.Get(pagePath)
	ce(err, "get")
	defer resp.Body.Close()
	var data struct {
		Code int
		Data struct {
			Upload_num       int
			Color            string
			Is_tx            int
			Discount_price   string
			Index_img_url    string
			Is_df            int
			Title            string // 标题
			Tid              int64
			Price            string
			Color_pics       string
			Id               int // 货号
			Art_no           string
			Imgs             string // 图片
			Support_services string // 退现 实拍 代发
			Discount_value   float64
			Shop_name        string
			Discount_type    string
			Attrs            string // 属性
			Is_sp            int
			Shop_id          int
			Size             string // 尺寸
			Bname            string // 市场名
			Up_time          string
			Bid              int
			Tcid             string
			Status           int
			Cid              string
			Desc             string // 描述html
		}
	}
	err = json.NewDecoder(resp.Body).Decode(&data)
	ce(err, "decode")

	dirName := fmt.Sprintf("%s-%s", time.Now().Format("2006-01-02-15-04-05"), os.Args[1])
	os.Mkdir(dirName, 0755)

	for i, imgPath := range strings.Split(data.Data.Imgs, ",") {
		// get image path
		if !strings.HasPrefix(imgPath, "http:") {
			imgPath = "http:" + imgPath
		}
		pt("%s\n", imgPath)
		// get image
		resp, err := http.Get(imgPath)
		ce(err, "get image")
		defer resp.Body.Close()
		// write to file
		fileName := filepath.Join(dirName, "foo-"+string([]byte{'a' + byte(i)})+
			path.Ext(imgPath))
		out, err := os.Create(fileName)
		ce(err, "create file")
		defer out.Close()
		//io.Copy(out, resp.Body)
		ce(vviccommon.CompositeLogo(resp.Body, out), "composite logo")
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(data.Data.Desc))
	ce(err, "goquery doc")
	doc.Find("img").Each(func(i int, se *goquery.Selection) {
		imgSrc, _ := se.Attr("src")
		pt("%s\n", imgSrc)
		fileName := filepath.Join(dirName, fmt.Sprintf("bar-%03d%s", i, path.Ext(imgSrc)))
		resp, err := http.Get(imgSrc)
		ce(err, "get image")
		defer resp.Body.Close()
		out, err := os.Create(fileName)
		ce(err, "create file")
		defer out.Close()
		io.Copy(out, resp.Body)
	})

	pt("\n")
	pt("梦丹铃 2016春%s\n", vviccommon.TidyTitle(data.Data.Title))
	pt("%s\n", data.Data.Discount_price)
	pt("%d\n", data.Data.Id)
	pt("\n")

	attrs := map[string]string{}
	for _, attr := range strings.Split(data.Data.Attrs, ",") {
		parts := strings.SplitN(attr, ":", 2)
		attrs[parts[0]] = parts[1]
	}
	attrKeys := []string{
		"风格",
		"裙长",
		"版型",
		"领型",
		"袖型",
		"元素",
		"颜色",
		"尺码",
		"图案",
		"适用",
		"组合",
		"款式",
		"袖长",
		"腰型",
		"门襟",
		"裙型",
		"质地",
	}
loop_key:
	for _, key := range attrKeys {
		for attrKey, attr := range attrs {
			if strings.Contains(attrKey, key) {
				pt("%-20s%s\n", attrKey, attr)
				delete(attrs, attrKey)
				continue loop_key
			}
		}
	}
	pt("\n")
	for attrKey, attr := range attrs {
		pt("%-20s%s\n", attrKey, attr)
	}

}

type Err struct {
	Pkg  string
	Info string
	Prev error
}

func (e *Err) Error() string {
	if e.Prev == nil {
		return fmt.Sprintf("%s: %s", e.Pkg, e.Info)
	}
	return fmt.Sprintf("%s: %s\n%v", e.Pkg, e.Info, e.Prev)
}

func me(err error, format string, args ...interface{}) *Err {
	if len(args) > 0 {
		return &Err{
			Pkg:  `vvicdownload`,
			Info: fmt.Sprintf(format, args...),
			Prev: err,
		}
	}
	return &Err{
		Pkg:  `vvicdownload`,
		Info: format,
		Prev: err,
	}
}

func ce(err error, format string, args ...interface{}) {
	if err != nil {
		panic(me(err, format, args...))
	}
}

func ct(err *error) {
	if p := recover(); p != nil {
		if e, ok := p.(error); ok {
			*err = e
		} else {
			panic(p)
		}
	}
}
