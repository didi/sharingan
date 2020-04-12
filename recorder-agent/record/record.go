package record

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/juju/ratelimit"
)

var (
	regex       *regexp.Regexp
	keywordsMap map[string][]*KeywordsBucket
	bucketMap   BucketMap
)

func init() {
	// init Regex
	regex = regexp.MustCompile(`(?U)"CallFromInbound".*"Request":"(?:GET|POST) (.*)[\s\\\?]`)

	// init keywordsMap
	keywordsMap = make(map[string][]*KeywordsBucket)
	for uri, keywordsRates := range keywordsMapConf {
		for _, keywordsRate := range keywordsRates {
			bucket := ratelimit.NewBucketWithRate(keywordsRate.rate, defaultBucketCapacity)
			keywordsBucket := &KeywordsBucket{keywordsRate.keywords, bucket}
			keywordsMap[uri] = append(keywordsMap[uri], keywordsBucket)
		}
	}

	// init bucketMap
	bucketMap.m = make(map[string]*ratelimit.Bucket)
	bucketMap.cap = bucketMapCapacity
	for uri, rate := range bucketMapConf {
		bucketMap.m[uri] = ratelimit.NewBucketWithRate(rate, defaultBucketCapacity)
	}
}

// Fliter 过滤记录
func Fliter(str string) (bool, error) {
	var err error

	uri := getURI(str)
	bucket := getBucket(str, uri)
	if bucket == nil {
		bucket, err = newBucket(uri)
	}

	if bucket != nil && bucket.TakeAvailable(1) > 0 {
		return false, nil
	}

	return true, err
}

// PrintBucketMap print bucket for debug
func PrintBucketMap() string {
	bucketMap.RLock()
	defer bucketMap.RUnlock()

	res := fmt.Sprintf("len:%d\n", len(bucketMap.m))
	res += "keys:\n"

	for k := range bucketMap.m {
		res += fmt.Sprintf("%s\n", k)
	}

	return res
}

func getURI(str string) string {
	uri := defaultURI

	matchs := regex.FindStringSubmatch(str)
	if len(matchs) >= 2 {
		uri = matchs[1]
	}

	return uri
}

func getBucket(str string, uri string) *ratelimit.Bucket {
	// uri + keywords 匹配
	if v, ok := keywordsMap[uri]; ok {
		for _, conf := range v {
			if matchKeyWords(str, conf.keywords) {
				return conf.bucket
			}
		}
	}

	// uri 匹配
	bucketMap.RLock()
	defer bucketMap.RUnlock()
	if b, ok := bucketMap.m[uri]; ok {
		return b
	}

	return nil
}

func matchKeyWords(str string, keywords []string) bool {
	for _, keyword := range keywords {
		if !strings.Contains(str, keyword) {
			return false
		}
	}

	return true
}

func newBucket(uri string) (*ratelimit.Bucket, error) {
	bucketMap.Lock()
	defer bucketMap.Unlock()

	// TODO sync DELETE
	cap := len(bucketMap.m)
	if cap >= bucketMap.cap {
		return nil, fmt.Errorf("bucketMap excess, len:%d", cap)
	}

	if _, ok := bucketMap.m[uri]; !ok {
		bucketMap.m[uri] = ratelimit.NewBucketWithRate(defaultBucketRate, defaultBucketCapacity)
	}

	return bucketMap.m[uri], nil
}
