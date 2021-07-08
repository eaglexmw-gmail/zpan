package provider

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/saltbo/gopkg/strutil"
	"github.com/upyun/go-sdk/v3/upyun"
)

// USSProvider 又拍云
type USSProvider struct {
	conf   *Config
	client *upyun.UpYun
}

func NewUSSProvider(conf *Config) (Provider, error) {
	return &USSProvider{
		conf: conf,
		client: upyun.NewUpYun(&upyun.UpYunConfig{
			Bucket:   conf.Bucket,
			Operator: conf.AccessKey,
			Password: conf.AccessSecret,
		}),
	}, nil
}

func (p *USSProvider) SetupCORS() error {
	// 官方没有提供相关接口，暂不实现
	return nil
}

func (p *USSProvider) List(prefix string) ([]Object, error) {
	panic("implement me")
}

func (p *USSProvider) Move(object, newObject string) error {
	panic("implement me")
}

func (p *USSProvider) SignedPutURL(key, filetype string, filesize int64, public bool) (url string, headers http.Header, err error) {
	//expireAt := time.Now().Add(time.Minute * 15).Unix()
	//headers.Set("X-Upyun-Expire", fmt.Sprint(expireAt))
	//headers.Set("X-Upyun-Uri-Prefix", uriPrefix)
	headers = make(http.Header)
	uri := fmt.Sprintf("/%s/%s", p.client.Bucket, key)
	date := time.Now().UTC().Format(http.TimeFormat)
	headers.Set("X-Date", date)
	headers.Set("Authorization", p.buildOldSign("PUT", uri, date, fmt.Sprint(filesize)))
	return fmt.Sprintf("http://v0.api.upyun.com/%s/%s", p.client.Bucket, key), headers, err
}

func (p *USSProvider) SignedGetURL(key, filename string) (url string, err error) {
	expireAt := time.Now().Add(time.Minute * 15).Unix()
	upt := p.buildUpt(expireAt, fmt.Sprintf("/%s", key))
	return fmt.Sprintf("%s?_upt=%s", p.PublicURL(key), upt), err
}

func (p *USSProvider) PublicURL(key string) (url string) {
	return fmt.Sprintf("%s/%s", p.conf.CustomHost, key)
}

func (p *USSProvider) ObjectDelete(key string) error {
	return p.client.Delete(&upyun.DeleteObjectConfig{
		Path:   key,
		Async:  false,
		Folder: false,
	})
}

func (p *USSProvider) ObjectsDelete(keys []string) error {
	for _, key := range keys {
		err := p.ObjectDelete(key)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *USSProvider) buildSign(items ...string) string {
	mac := hmac.New(sha1.New, []byte(strutil.Md5Hex(p.conf.AccessSecret)))
	mac.Write([]byte(strings.Join(items, "&")))
	signStr := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	return fmt.Sprintf("UpYun %s:%s", p.client.Operator, signStr)
}

func (p USSProvider) buildUpt(expireAt int64, uri string) string {
	// sign = MD5( secret & etime & URI )
	//_upt = sign { 中间 8 位 }＋etime
	signStr := strings.Join([]string{p.conf.AccessSecret, fmt.Sprint(expireAt), uri}, "&")
	return strutil.Md5Hex(signStr)[12:20] + fmt.Sprint(expireAt)
}

func (p *USSProvider) buildOldSign(items ...string) string {
	items = append(items, strutil.Md5Hex(p.conf.AccessSecret))
	signStr := strutil.Md5Hex(strings.Join(items, "&"))
	return fmt.Sprintf("UpYun %s:%s", p.client.Operator, signStr)
}
