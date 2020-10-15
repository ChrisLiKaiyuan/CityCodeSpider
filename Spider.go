package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/parnurzeal/gorequest"
	"os"
	"regexp"
)

func main() {

	file, err := os.OpenFile("data.json", os.O_CREATE|os.O_WRONLY, 0)
	if err != nil {
		fmt.Println("Open File Failed.")
		return
	}

	request := gorequest.New()
	_, body, errors := request.Get("http://www.mca.gov.cn//article/sj/xzqh/2020/2020/2020092500801.html").End()
	if errors != nil {
		fmt.Println("Request Failed.")
		return
	}

	codeRe := regexp.MustCompile(`<td class=xl[\d]{7}>([\d]{6})</td>|<td class=xl[\d]{7}><span lang=EN-US>([\d]{6})</span></td>`)
	cityRe := regexp.MustCompile(`<td class=xl[\d]{7}><span style='mso-spacerun:yes'>.+?</span>([\D]{1,100})<span[.\s]+?style='mso-spacerun:yes'>.+?</span></td>|<td class=xl[\d]{7}><span style='mso-spacerun:yes'>.+?</span>([\D]{1,100})</td>|<td class=xl[\d]{7}>([\D]{1,100})</td>`)
	codeRaw := codeRe.FindAllStringSubmatch(body, -1)
	cityRaw := cityRe.FindAllStringSubmatch(body, -1)

	var CodeCityMap map[string]string
	CodeCityMap = make(map[string]string)
	code := ""
	city := ""

	length := len(codeRaw)
	for index := 0; index < length; index++ {
		if codeRaw[index][1] != "" {
			code = codeRaw[index][1]
		} else {
			code = codeRaw[index][2]
		}
		if cityRaw[index][1] != "" {
			city = cityRaw[index][1]
		} else if cityRaw[index][2] != "" {
			city = cityRaw[index][2]
		} else {
			city = cityRaw[index][3]
		}
		CodeCityMap[code] = city
	}

	type Json struct {
		Code string
		Info interface{}
	}
	type JsonInfo struct {
		Name          string
		LevelProvince string
		LevelCity     string
		LevelArea     string
	}

	var dataz map[string]Json
	dataz = make(map[string]Json)
	for mapCode := range CodeCityMap {
		codeProvince := mapCode[0:2] + "0000"
		codeCity := mapCode[0:4] + "00"
		codeArea := mapCode
		province, _ := CodeCityMap[codeProvince]
		city, ok := CodeCityMap[codeCity]
		area, _ := CodeCityMap[codeArea]

		if !ok {
			city = ""
			codeCity = codeArea
		}
		if mapCode[2:6] == "0000" {
			city = ""
			area = ""
		}
		if mapCode[4:6] == "00" {
			area = ""
		}

		dataz[mapCode] = Json{
			Code: mapCode,
			Info: JsonInfo{
				Name:          province + city + area,
				LevelProvince: codeProvince,
				LevelCity:     codeCity,
				LevelArea:     codeArea,
			},
		}
	}

	data, e := json.Marshal(dataz)
	if e != nil {
		fmt.Println("Convert Failed.")
		return
	}
	var buf bytes.Buffer
	_ = json.Indent(&buf, data, "", "    ")
	_, _ = fmt.Fprintln(file, buf.String())

	_ = file.Close()
}
