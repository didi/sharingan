package record

import (
	"sync"

	"github.com/juju/ratelimit"
)

const defaultBucketCapacity = 1   // 默认bucket容量
const defaultBucketRate = 0.2     // 默认bucket速率，0.2r/s
const defaultURI = "/api/default" // 默认请求URI
const bucketMapCapacity = 1000    // BucketMap容量

// KeywordsRate KeywordsRate
type KeywordsRate struct {
	keywords []string
	rate     float64
}

// KeywordsBucket KeywordsBucket
type KeywordsBucket struct {
	keywords []string
	bucket   *ratelimit.Bucket
}

var keywordsMapConf = map[string][]*KeywordsRate{
	"/api/v1/get": []*KeywordsRate{
		&KeywordsRate{[]string{`\"param_a\":\"xxxx\"`}, 1},
	},
}

// BucketMap BucketMap
type BucketMap struct {
	sync.RWMutex
	m   map[string]*ratelimit.Bucket
	cap int
}

var bucketMapConf = map[string]float64{
	"/api/default": defaultBucketRate,
}
