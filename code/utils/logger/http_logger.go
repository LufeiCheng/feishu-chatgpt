package logger

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

// HttpLogger is a logger that send log to http server
type httpLogger struct {
	url    string
	method string
	// log cache for batch send
	logCache *[]string
	// last send timestamp
	lastSendTime int64
	// log send interval
	interval int
	// log send threshold
	threshold int
}

// NewHttpLogger create a new http logger
func NewHttpLogger(url, method string, interval, threshold int) *httpLogger {
	return &httpLogger{
		url:          url,
		method:       method,
		logCache:     &[]string{},
		lastSendTime: time.Now().Unix(),
		interval:     interval,
		threshold:    threshold,
	}
}

// Log send log to http server
func (l *httpLogger) Log(log string) {
	*l.logCache = append(*l.logCache, log)
	if len(*l.logCache) >= l.threshold {
		l.send()
	} else if time.Now().Unix()-l.lastSendTime >= int64(l.interval) {
		l.send()
	}
}

// send send log to http server
func (l *httpLogger) send() {
	body := strings.NewReader(strings.Join(*l.logCache, "----"))
	req, err := http.NewRequest(l.method, l.url, body)
	if err != nil {
		fmt.Println("请求初始化失败", err)
		return
	}
	req.Header.Set("Content-Type", "text/plain")

	// send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("发送请求失败", err)
		return
	}
	defer resp.Body.Close()

	// reset cache
	*l.logCache = []string{}
	// reset last send timestamp
	l.lastSendTime = time.Now().Unix()

	fmt.Println("response Status:", resp.Status)
}

// implement Write interface
func (l *httpLogger) Write(p []byte) (n int, err error) {
	l.Log(string(p))
	return len(p), nil
}

// implement Flush interface
func (l *httpLogger) Flush() {
	l.send()
}

// Close close the logger
func (l *httpLogger) Close() {
	l.send()
}
