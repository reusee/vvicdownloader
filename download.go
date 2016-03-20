package main

import (
	"encoding/json"
	"fmt"
	"io"
	"path"
	"path/filepath"
	"strings"

	"github.com/PuerkitoBio/goquery"

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

	dirName := fmt.Sprintf("id-%s", os.Args[1])
	os.Mkdir(dirName, 0755)

	for i, imgPath := range strings.Split(data.Data.Imgs, ",") {
		if !strings.HasPrefix(imgPath, "http:") {
			imgPath = "http:" + imgPath
		}
		pt("%s\n", imgPath)
		fileName := filepath.Join(dirName, string([]byte{'a' + byte(i)})+
			path.Ext(imgPath))
		resp, err := http.Get(imgPath)
		ce(err, "get image")
		defer resp.Body.Close()
		out, err := os.Create(fileName)
		ce(err, "create file")
		defer out.Close()
		io.Copy(out, resp.Body)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(data.Data.Desc))
	ce(err, "goquery doc")
	doc.Find("img").Each(func(i int, se *goquery.Selection) {
		imgSrc, _ := se.Attr("src")
		pt("%s\n", imgSrc)
	})
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
