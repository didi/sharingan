package match

var (
	matchers []Matcher
)

type Matcher interface {
	Ignore(request []byte) bool
}

type FuncMatcher func(request []byte) bool

func (m FuncMatcher) Ignore(request []byte) bool {
	return m(request)
}

// AddMatcher 添加一个匹配器
func AddMatcher(m Matcher) {
	matchers = append(matchers, m)
}

// Ignored 判断一段数据是否需要忽略
func Ignored(request []byte) bool {
	for _, m := range matchers {
		if m.Ignore(request) {
			return true
		}
	}
	return false
}
