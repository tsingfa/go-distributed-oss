package es

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type Metadata struct {
	Name    string
	Version int
	Size    int64
	Hash    string
}

type hit struct {
	Source Metadata `json:"_source"`
}

type searchResult struct {
	Hits struct {
		Total struct {
			Value    int
			Relation string
		}
		Hits []hit
	}
}

// getMetadata 根据对象的名字和版本号，获取指定版本的对象的元数据
func getMetadata(name string, version int) (meta Metadata, err error) {
	urlStr := fmt.Sprintf("http://%s/metadata/_doc/%s_%d/_source",
		os.Getenv("ES_SERVER"), name, version) //索引为metadata,类型为objects,文档id由name和version拼接而成
	resp, err := http.Get(urlStr) //向ES服务器请求，这里确切指定了name_version，不用再按条件搜索
	if err != nil {
		log.Println(err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("fail to get %s_%d from %s: %d", name, version, os.Getenv("ES_SERVER"), resp.StatusCode)
		log.Println(err)
		return
	}
	result, _ := io.ReadAll(resp.Body)
	_ = json.Unmarshal(result, &meta) //将json格式的result中的数据，解析到meta结构体中
	return
}

// SearchLatestVersion 根据对象名，获取其最新版的元数据
func SearchLatestVersion(name string) (meta Metadata, err error) {
	urlStr := fmt.Sprintf("http://%s/metadata/_search?q=name:%s&size=1&sort=version:desc",
		os.Getenv("ES_SERVER"), url.PathEscape(name)) //有name而没有version，需要调用ES的搜索API
	resp, err := http.Get(urlStr)
	if err != nil {
		return
	}
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("fail to search latest metadata: %d", resp.StatusCode)
		return
	}
	result, _ := io.ReadAll(resp.Body)
	var sr searchResult //适配搜索结果的结构体
	_ = json.Unmarshal(result, &sr)
	if len(sr.Hits.Hits) != 0 {
		meta = sr.Hits.Hits[0].Source //如果有搜索到，返回降序结果的第一条
	}
	return
}

// GetMetadata 根据对象名和版本号，返回指定版本的元数据。
// 如果未指定版本号（为0），则返回最新版。
func GetMetadata(name string, version int) (Metadata, error) {
	if version == 0 { //若没有指定版本（为0），则获取最新版
		return SearchLatestVersion(name)
	}
	return getMetadata(name, version)
}

// PutMetadata 向ES服务器上传一个新的元数据
func PutMetadata(name string, version int, size int64, hash string) error {
	doc := fmt.Sprintf(`{"name":"%s","version":%d,"size":%d,"hash":"%s"}`,
		name, version, size, hash)
	client := http.Client{}
	urlStr := fmt.Sprintf("http://%s/metadata/_doc/%s_%d?op_type=create",
		os.Getenv("ES_SERVER"), name, version)
	request, _ := http.NewRequest("PUT", urlStr, strings.NewReader(doc))
	request.Header.Set("Content-Type", "application/json") // 添加Content-Type头部
	resp, err := client.Do(request)
	//log.Println("resp:\n", resp)
	if err != nil {
		log.Println("ES PutMetadata error:", err)
		return err
	}
	if resp.StatusCode == http.StatusConflict { //如果同时有多个客户端上传同一个元数据（文档id相冲突）
		return PutMetadata(name, version+1, size, hash) //则第一个客户端会成功，而失败的版本+1后重试，直至都不冲突
	}
	if resp.StatusCode != http.StatusCreated {
		result, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("fail to put metadata: %d %s", resp.StatusCode, string(result))
	}
	return nil
}

// AddVersion 令指定对象的最新版本记录+1
func AddVersion(name, hash string, size int64) error {
	version, err := SearchLatestVersion(name)
	//log.Println("version:", version)
	if err != nil {
		return err
	}
	return PutMetadata(name, version.Version+1, size, hash)
}

// SearchAllVersions 搜索某个对象或所有对象的全部版本，返回元数据数组。
// from和size参数指定分页的显示结果。
func SearchAllVersions(name string, from, size int) ([]Metadata, error) {
	urlStr := fmt.Sprintf("http://%s/metadata/_search?sort=name,version&from=%d&size=%d",
		os.Getenv("ES_SERVER"), from, size)
	if name != "" { //name不为空，即有指定的对象，增加查询条件
		urlStr += "&q=name:" + name
	}
	resp, err := http.Get(urlStr)
	if err != nil {
		return nil, err
	}
	metas := make([]Metadata, 0)
	result, _ := io.ReadAll(resp.Body)
	var sr searchResult
	_ = json.Unmarshal(result, &sr)
	for i := range sr.Hits.Hits {
		metas = append(metas, sr.Hits.Hits[i].Source)
	}
	return metas, nil
}

func DelMetadata(name string, version int) {
	client := http.Client{}
	urlStr := fmt.Sprintf("http://%s/metadata/_doc/%s_%d",
		os.Getenv("ES_SERVER"), name, version)
	request, _ := http.NewRequest("DELETE", urlStr, nil)
	_, _ = client.Do(request)
}

type Bucket struct {
	Key        string
	DocCount   int
	MinVersion struct {
		Value float32
	}
}

type aggregateResult struct {
	Aggregations struct {
		GroupByName struct {
			Buckets []Bucket
		}
	}
}

func SearchVersionStatus(minDocCount int) ([]Bucket, error) {
	client := http.Client{}
	urlStr := fmt.Sprintf("http://%s/metadata/_search", os.Getenv("ES_SERVER"))
	body := fmt.Sprintf(`
        {
          "size": 0,
          "aggs": {
            "group_by_name": {
              "terms": {
                "field": "name",
                "min_doc_count": %d
              },
              "aggs": {
                "min_version": {
                  "min": {
                    "field": "version"
                  }
                }
              }
            }
          }
        }`, minDocCount)
	req, _ := http.NewRequest("GET", urlStr, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json") // 添加Content-Type头部
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	b, _ := io.ReadAll(resp.Body)
	var ar aggregateResult
	_ = json.Unmarshal(b, &ar)
	return ar.Aggregations.GroupByName.Buckets, nil
}

func HasHash(hash string) (bool, error) {
	urlStr := fmt.Sprintf("http://%s/metadata/_search?q=hash:%s&size=0", os.Getenv("ES_SERVER"), hash)
	r, e := http.Get(urlStr)
	if e != nil {
		return false, e
	}
	b, _ := io.ReadAll(r.Body)
	var sr searchResult
	_ = json.Unmarshal(b, &sr)
	return sr.Hits.Total.Value != 0, nil
}

func SearchHashSize(hash string) (size int64, e error) {
	urlStr := fmt.Sprintf("http://%s/metadata/_search?q=hash:%s&size=1",
		os.Getenv("ES_SERVER"), hash)
	r, e := http.Get(urlStr)
	if e != nil {
		return
	}
	if r.StatusCode != http.StatusOK {
		e = fmt.Errorf("fail to search hash size: %d", r.StatusCode)
		return
	}
	result, _ := io.ReadAll(r.Body)
	var sr searchResult
	_ = json.Unmarshal(result, &sr)
	if len(sr.Hits.Hits) != 0 {
		size = sr.Hits.Hits[0].Source.Size
	}
	return
}
