package conf

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/url"
	"regexp"
	"strings"
)
// config

type MatchExp struct {
	Query  string
	Attr   string
	Folder interface{} //可选值url,title,none,正则表达式
}

type Proxy struct {
	Server   string
	UserName string
	Password string
}

type Config struct {
	Root         *url.URL
	Proxy        *Proxy
	ImageExp   []*MatchExp
	PageRegex    []*regexp.Regexp
	ImgPageRegex []*regexp.Regexp
	HrefExp    []*MatchExp
	Header       map[string]string
	Charset      string
}

func (c *Config) Load(file string) error {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	content = bytes.Replace(content, []byte("\\"), []byte("\\\\"), -1)
	content = bytes.Replace(content, []byte("\\\\\""), []byte("\\\""), -1)

	comRegex := regexp.MustCompile(`\s*/\*.*\*/`)
	content = comRegex.ReplaceAll(content, []byte{}) //删除注释

	nRegex := regexp.MustCompile(`\n|\t|\r`)
	content = nRegex.ReplaceAll(content, []byte{})

	jsonObj := make(map[string]interface{})
	err = json.Unmarshal(content, &jsonObj)
	if err != nil {
		log.Println(string(content))
		return errors.New("[1]配置文件格式有误:" + err.Error())
	}

	root := jsonObj["root"].(string)
	temp := strings.ToLower(root)
	if !strings.HasPrefix(temp, "http://") && !strings.HasPrefix(temp, "https://") {
		root = "http://" + root
	}

	c.Root, err = url.Parse(root)
	if err != nil {
		return err
	}

	//代理服务器配置
	c.Proxy = &Proxy{}
	proxy, ok := jsonObj["proxy"].(map[string]interface{})
	if !ok {
		return errors.New("[2]解析proxy时出错")
	}
	c.Proxy.Server = proxy["server"].(string)
	c.Proxy.UserName = proxy["username"].(string)
	c.Proxy.Password = proxy["password"].(string)

	//HTTP请求头
	headers, ok := jsonObj["header"].(map[string]interface{})
	if !ok {
		return errors.New("[3]解析header时出错")
	}
	c.Header = make(map[string]string)
	for k, v := range headers {
		c.Header[k] = v.(string)
	}

	//网页编码
	c.Charset, ok = jsonObj["charset"].(string)
	if !ok {
		return errors.New("解析charset时出错")
	}
	if (strings.TrimSpace(c.Charset) == "") {
		c.Charset = "utf-8"
	}

	//正则表达式
	reg, ok := jsonObj["regex"].(map[string]interface{})
	if !ok {
		return errors.New("[4]解析regex时出错")
	}

	imgRegs, ok := reg["image"].([]interface{})
	if ok {
		c.ImageExp = make([]*MatchExp, len(imgRegs))
		for i, val := range imgRegs {
			obj, ok := val.(map[string]interface{})
			if !ok {
				return errors.New("[5]解析regex.image时出错")
			}
			c.ImageExp[i] = &MatchExp{}
			c.ImageExp[i].Query = obj["query"].(string)
			c.ImageExp[i].Attr = obj["attr"].(string)

			folder := strings.ToLower(obj["folder"].(string))
			if folder != "none" && folder != "url" && folder != "title" {
				c.ImageExp[i].Folder, err = regexp.Compile(folder)
				if err != nil {
					return errors.New("[6]解析正则表达式" + folder + "时出错")
				}
			} else {
				c.ImageExp[i].Folder = folder
			}
		}
	} else {
		return errors.New("[8]解析regex.image时出错")
	}

	pageRegs, ok := reg["page"].([]interface{})
	if ok {
		c.PageRegex = make([]*regexp.Regexp, len(pageRegs))
		for i, val := range pageRegs {
			valStr := val.(string)
			c.PageRegex[i], err = regexp.Compile(valStr)
			if err != nil {
				return errors.New("[9]解析正则表达式" + valStr + "时出错")
			}
		}
	} else {
		return errors.New("[10]解析regex.page时出错")
	}

	imgPageRegs, ok := reg["imgInPage"].([]interface{})
	if ok {
		c.ImgPageRegex = make([]*regexp.Regexp, len(imgPageRegs))
		for i, val := range imgPageRegs {
			valStr := val.(string)
			c.ImgPageRegex[i], err = regexp.Compile(valStr)
			if err != nil {
				return errors.New("[11]解析正则表达式" + valStr + "时出错")
			}
		}
	} else {
		return errors.New("[12]解析regex.imgInPage时出错")
	}

	hrefRegex, ok := reg["href"].([]interface{})
	if ok {
		c.HrefExp = make([]*MatchExp, len(hrefRegex))
		for i, val := range hrefRegex {
			obj, ok := val.(map[string]interface{})
			if !ok {
				return errors.New("[13]解析regex.href时出错")
			}
			c.HrefExp[i] = &MatchExp{}
			c.HrefExp[i].Query = obj["query"].(string)
			c.HrefExp[i].Attr = obj["attr"].(string)
		}
	} else {
		return errors.New("[15]解析regex.hrefRegex时出错")
	}

	return nil
}
