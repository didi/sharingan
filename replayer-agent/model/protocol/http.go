package protocol

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/didichuxing/sharingan/replayer-agent/utils/helper"
)

var textprotoReaderPool sync.Pool

func newTextprotoReader(br *bufio.Reader) *textproto.Reader {
	if v := textprotoReaderPool.Get(); v != nil {
		tr := v.(*textproto.Reader)
		tr.R = br
		return tr
	}
	return textproto.NewReader(br)
}

func putTextprotoReader(r *textproto.Reader) {
	r.R = nil
	textprotoReaderPool.Put(r)
}

func ParseHTTP(body string) (pairs map[string]json.RawMessage, requestMark string, err error) {
	reader := bufio.NewReaderSize(bytes.NewBuffer(helper.StringToBytes(body)), len(body))

	tp := newTextprotoReader(reader)
	defer func() {
		putTextprotoReader(tp)
	}()

	// First line: GET /index.html HTTP/1.0
	var s string
	if s, err = tp.ReadLine(); err != nil {
		return nil, "", err
	}

	method, URI, _, ok := parseLine(s)
	if !ok {
		return nil, "", errors.New("malformed HTTP request")
	}

	URL, err := url.ParseRequestURI(URI)
	if err == nil {
		requestMark = URL.Path
	}

	// Subsequent lines: Key: value.
	mimeHeaders, err := tp.ReadMIMEHeader()
	if err != nil {
		return nil, "", err
	}

	var params []byte
	if method == "GET" {
		params, _ = json.Marshal(URL.Query())
	} else {
		params, _ = parseBody(reader, mimeHeaders)
	}

	pairs, err = helper.Json2SingleLayerMap(params)
	// 添加关注的Header
	addHeaders(pairs, mimeHeaders)

	return pairs, requestMark, nil
}

// SortHTTP 对http参数排序，提高匹配成功率
func SortHTTP(body []byte) (sortBody []byte, err error) {
	reader := bufio.NewReaderSize(bytes.NewBuffer(body), len(body))

	tp := newTextprotoReader(reader)
	defer func() {
		putTextprotoReader(tp)
	}()

	// First line: GET /index.html HTTP/1.0
	var s string
	if s, err = tp.ReadLine(); err != nil {
		return body, err
	}

	method, URI, _, ok := parseLine(s)
	if !ok {
		return body, errors.New("malformed HTTP request")
	}

	URL, err := url.ParseRequestURI(URI)
	if err != nil {
		return body, err
	}

	// 对 URL.RawQuery 排序，并替换body
	qs := URL.Query()
	sortedQuery := sortQuery(qs)
	if sortedQuery != URL.RawQuery {
		body = []byte(strings.Replace(string(body), URL.RawQuery, sortedQuery, 1))
	}

	// GET 则排序结束
	if method == "GET" {
		return body, nil
	}

	// Subsequent lines: Key: value.
	mimeHeaders, err := tp.ReadMIMEHeader()
	if err != nil {
		return body, err
	}

	ct := mimeHeaders.Get("Content-Type")
	if !strings.Contains(ct, "application/x-www-form-urlencoded") {
		return body, nil
	}

	// 解析 application/x-www-form-urlencoded 下body内的参数，并排序
	query, qs, err := parseUrlencode(reader)
	if err != nil {
		return body, err
	}
	sortedQuery = sortQuery(qs)
	if sortedQuery != query {
		body = []byte(strings.Replace(string(body), query, sortedQuery, 1))
	}

	return body, nil
}

// sortQuery 对参数排序, 其中qs的值经过url.QueryUnescape()
func sortQuery(qs url.Values) string {
	encodeSorted := ""
	keys := make([]string, 0)
	for key, _ := range qs {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		vs := qs[key]
		encodeSorted = encodeSorted + url.QueryEscape(key) + "=" + url.QueryEscape(vs[len(vs)-1]) + "&"
	}
	encodeSorted = strings.TrimSuffix(encodeSorted, "&")

	return encodeSorted
}

// parseRequestLine parses "GET /foo HTTP/1.1" into its three parts.
func parseLine(line string) (method, requestURI, proto string, ok bool) {
	s1 := strings.Index(line, " ")
	s2 := strings.Index(line[s1+1:], " ")
	if s1 < 0 || s2 < 0 {
		return
	}
	s2 += s1 + 1
	return line[:s1], line[s1+1 : s2], line[s2+1:], true
}

func parseBody(r *bufio.Reader, headers textproto.MIMEHeader) ([]byte, error) {
	te := headers.Get("Transfer-Encoding")
	if strings.Contains(te, "chunked") {
		r = bufio.NewReaderSize(NewChunkedReader(r), r.Buffered())
	}

	b, err := r.Peek(1)
	if err != nil {
		return nil, err
	}

	// fast return
	if b[0] == '{' || b[0] == '[' {
		return r.ReadBytes('\r')
	}

	ct := headers.Get("Content-Type")
	if strings.Contains(ct, "multipart/form-data") {
		return parseMultiPart(ct, r)
	}

	// treat body as form by default
	return parseForm(ct, r)
}

func parseMultiPart(ct string, r *bufio.Reader) ([]byte, error) {
	d, params, err := mime.ParseMediaType(ct)
	if err != nil || d != "multipart/form-data" {
		return nil, http.ErrNotMultipart
	}
	boundary, ok := params["boundary"]
	if !ok {
		return nil, http.ErrMissingBoundary
	}

	res := make(map[string][]interface{})

	r2 := multipart.NewReader(r, boundary)
	for true {
		p, err := r2.NextPart()
		if err != nil {
			break
		}

		buf := make([]byte, r.Buffered())
		n, err := p.Read(buf)
		if err != nil {
			continue
		}

		if formname := p.FormName(); formname != "" {
			res[formname] = append(res[formname], buf[:n])
		}
		if filename := p.FileName(); filename != "" {
			// avoid random string for filename
			res["filename"] = append(res["filename"], filename)
		}
	}

	return json.Marshal(res)
}

func parseForm(ct string, r *bufio.Reader) ([]byte, error) {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	vs, err := url.ParseQuery(helper.BytesToString(b))
	if err != nil {
		return nil, err
	}
	return json.Marshal(vs)
}

func parseUrlencode(r *bufio.Reader) (string, url.Values, error) {
	var query string
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return query, nil, err
	}

	query = helper.BytesToString(b)
	vs, err := url.ParseQuery(query)
	if err != nil {
		return query, nil, err
	}
	return query, vs, nil
}

func addHeaders(pairs map[string]json.RawMessage, headers textproto.MIMEHeader) {
	interests := []string{
		"Content-Type",
		"Origin",
		"Access-Control-Allow-Origin",
		"Access-Control-Allow-Credentials",
		"Access-Control-Allow-Methods",
		"Access-Control-Allow-Headers",
	}

	for _, i := range interests {
		if v := headers.Get(i); len(v) > 0 {
			pairs[i] = json.RawMessage(strconv.Quote(v))
		}
	}
}
