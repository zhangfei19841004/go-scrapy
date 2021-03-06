package scrapy

import (
	"github.com/Genesis-Palace/requests"
	"github.com/ymzuiku/hit"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type Proxy bool

const (
	GET      = "get"
	POST     = "post"
	POSTJSON = "post-json"
	// 使用代理时 setProxy(Use)
	Use            Proxy = true
	DefaultTimeOut       = time.Second
)

func NewAbutunProxy(appid, secret, proxyServer string) *AbuyunProxy {
	return &AbuyunProxy{
		AppID:       appid,
		AppSecret:   secret,
		ProxyServer: proxyServer,
	}
}

type AbuyunProxy struct {
	AppID       string
	AppSecret   string
	ProxyServer string
}

func (p AbuyunProxy) ProxyClient() *http.Client {
	proxyUrl, _ := url.Parse("http://" + p.AppID + ":" + p.AppSecret + "@" + p.ProxyServer)
	return &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyUrl)}}
}

type ClientInterface interface {
	Get(url String, args ...interface{}) (*requests.Response, error)
	PostJson(url, js String, args ...interface{}) (*requests.Response, error)
	SetTimeOut(duration time.Duration)
	SetHeaders(header requests.Header)
}

type DefaultClient struct {
	c *requests.Request
	sync.RWMutex
}

func (d *DefaultClient) SetTimeOut(duration time.Duration) {
	d.Lock()
	defer d.Unlock()
	d.c.SetTimeout(duration)
}
func (d *DefaultClient) Get(url String, args ...interface{}) (resp *requests.Response, err error) {
	d.Lock()
	defer d.Unlock()
	return d.c.Get(url.String(), args)
}

func (d *DefaultClient) PostJson(url, js String, args ...interface{}) (*requests.Response, error) {
	d.Lock()
	defer d.Unlock()
	return d.c.PostJson(url.String(), js.String(), args)
}

func (d *DefaultClient) SetHeaders(header requests.Header) {
	d.Lock()
	defer d.Unlock()
	for k, v := range header {
		d.c.Header.Add(k, v)
	}
}

type ProxyClient struct {
	c *requests.Request
	sync.RWMutex
}

func (p *ProxyClient) SetHeaders(header requests.Header) {
	p.Lock()
	defer p.Unlock()
	for k, v := range header {
		p.c.Header.Add(k, v)
	}
}

func (p *ProxyClient) SetTimeOut(duration time.Duration) {
	p.Lock()
	defer p.Unlock()
	p.c.SetTimeout(duration)
}

func (p *ProxyClient) Get(url String, args ...interface{}) (resp *requests.Response, err error) {
	p.Lock()
	defer p.Unlock()
	return p.c.Get(url.String(), args)
}

func (p *ProxyClient) PostJson(url, js String, args ...interface{}) (*requests.Response, error) {
	p.Lock()
	defer p.Unlock()
	return p.c.PostJson(url.String(), js.String(), args)
}

type Requests struct {
	Url     String
	headers requests.Header
	cookies *http.Cookie
	proxy   bool
	method  String
	timeout time.Duration
	json    String
	c       ClientInterface
	sync.RWMutex
}

func (r *Requests) Json(js String) *Requests {
	r.Lock()
	r.json = js
	r.Unlock()
	return r
}

func (r *Requests) SetMethod(method string) *Requests {
	r.Lock()
	r.method = String(method)
	r.Unlock()
	return r
}

func (r *Requests) SetTimeOut(timeout time.Duration) *Requests {
	r.Lock()
	r.timeout = timeout
	r.c.SetTimeOut(timeout)
	r.Unlock()
	return r
}

func (r *Requests) SetHeader(headers requests.Header) *Requests {
	r.Lock()
	r.Unlock()
	r.headers = headers
	r.c.SetHeaders(headers)
	return r
}

func (r *Requests) SetCookies(cookie *http.Cookie) *Requests {
	r.Lock()
	r.cookies = cookie
	r.Unlock()
	return r
}

func (r *Requests) timeoutIsNil() bool {
	return r.timeout.Microseconds() == 0
}

func (r *Requests) Do() (resp *requests.Response, err error) {
	if r.method.Empty() {
		r.method = GET
	}
	durtion := hit.If(r.timeout > DefaultTimeOut, r.timeout, DefaultTimeOut).(time.Duration)
	r.c.SetTimeOut(durtion)
	switch r.method {
	case GET:
		return r.c.Get(r.Url)
	case POSTJSON:
		return r.c.PostJson(r.Url, r.json)
	case POST:
	}
	panic("unreach")
}

type Response struct {
	*requests.Response
}

func NewRequest(url String, args ...interface{}) *Requests {
	var req = &Requests{
		Url:     url,
		RWMutex: sync.RWMutex{},
		c:       NewDefaultClient(),
	}
	for _, arg := range args {
		switch arg.(type) {
		case requests.Header:
			req.SetHeader(arg.(requests.Header))
		case *http.Cookie:
			req.SetCookies(arg.(*http.Cookie))
		case time.Duration:
			req.SetTimeOut(arg.(time.Duration))
		case *AbuyunProxy:
			req.c = NewProxyClient(arg.(*AbuyunProxy))
		default:
		}
	}
	return req
}

func NewDefaultClient() ClientInterface {
	return &DefaultClient{
		c:       requests.Requests(),
		RWMutex: sync.RWMutex{},
	}
}

func NewProxyClient(proxy *AbuyunProxy) ClientInterface {
	client := requests.Requests()
	client.Client = proxy.ProxyClient()
	return &ProxyClient{
		c:       client,
		RWMutex: sync.RWMutex{},
	}
}
