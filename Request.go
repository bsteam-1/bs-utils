package bs_utils

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/pkg/errors"
)

type Req struct {
	c                 *resty.Client
	Debug             bool
	REQUEST_STATE_log Logs
	Log_Path          string
}

func (r *Req) Init() {
	r.c = resty.New()
	if r.Log_Path != "" {
		r.REQUEST_STATE_log.Init()
	}
	r.c.AddRetryCondition(func(response *resty.Response, e error) bool {
		return response.StatusCode() == http.StatusBadRequest
	})
	r.c.SetHeader("Connection", "Keep-Alive")
	if r.Debug {
		r.c.EnableTrace()
	}
}
func (r *Req) Get(url string, timeout time.Duration, params map[string]string, header map[string]string, retry bool) (str string, err error) {
	defer func() {
		s := recover()
		if s != nil {
			err = errors.Errorf("%v", s)
		}
	}()
	if retry {
		r.c.SetRetryCount(3).SetRetryMaxWaitTime(2 * time.Second)
	} else {
		r.c.SetRetryCount(0)
	}
	r.c.SetHeaders(header)
	r.c.SetTimeout(timeout)
	rd, e := r.c.R().SetQueryParams(params).Get(url)
	r.debug(rd)
	if e != nil {
		if r.Log_Path != "" {
			r.REQUEST_STATE_log.sugar.Errorw(Join("GET ", e.Error()))
		}
		return "", e
	}
	return rd.String(), nil
}
func (r *Req) Post(url string, timeout time.Duration, params url.Values, header map[string]string, retry bool) (str string, err error) {
	defer func() {
		s := recover()
		if s != nil {
			str = ""
			err = errors.Errorf("%v", s)
		}
	}()
	if retry {
		r.c.SetRetryCount(3).SetRetryMaxWaitTime(2 * time.Second)
	} else {
		r.c.SetRetryCount(0)
	}
	r.c.SetHeaders(header)
	r.c.SetHeader("Content-Type", "application/x-www-form-urlencoded")
	r.c.SetTimeout(timeout)
	rd, e := r.c.R().SetBody(params.Encode()).Post(url)
	r.debug(rd)
	if e != nil {
		if r.Log_Path != "" {
			r.REQUEST_STATE_log.sugar.Errorw(Join("POST ", e.Error()))
		}
		return "", e
	}
	return rd.String(), nil
}

func (r *Req) Raw(url string, retry bool, timeout time.Duration) (data []byte) {
	defer func() {
		s := recover()
		if s != nil {
			P(fmt.Sprintf("%v", s))
			data = nil
		}
	}()
	if retry {
		r.c.SetRetryCount(3).SetRetryMaxWaitTime(2 * time.Second)
	} else {
		r.c.SetRetryCount(0)
	}

	r.c.SetTimeout(timeout)
	rd, e := r.c.R().EnableTrace().Get(url)
	if e != nil {
		if r.Log_Path != "" {
			r.REQUEST_STATE_log.sugar.Errorw(Join("Raw ", e.Error()))
		}
		return nil
	} else {
		return rd.Body()
	}
}
func (r *Req) Upload(Filepath, SaveFolder string, timeout time.Duration) (err error) {
	defer func() {
		s := recover()
		if s != nil {
			err = errors.Errorf("%v", s)
		}
	}()
	r.c.SetTimeout(timeout)
	_, e := r.c.R().SetFile("file", Filepath).SetFormData(map[string]string{"dir": SaveFolder}).
		Post("https://nas.brainftp.dev:6666/upload")
	if e != nil {
		r.REQUEST_STATE_log.sugar.Errorw(Join("Upload ", e.Error()))
		return e
	} else {
		return nil
	}
}
func (r *Req) debug(rd *resty.Response) {
	if r.Debug {
		ti := rd.Request.TraceInfo()
		fmt.Println("DNSLookup    :", ti.DNSLookup)
		fmt.Println("ConnTime     :", ti.ConnTime)
		fmt.Println("TCPConnTime  :", ti.TCPConnTime)
		fmt.Println("TLSHandshake :", ti.TLSHandshake)
		fmt.Println("ServerTime   :", ti.ServerTime)
		fmt.Println("ResponseTime :", ti.ResponseTime)
		fmt.Println("IsConnReused :", ti.IsConnReused)
		fmt.Println("IsConnWasIdle:", ti.IsConnWasIdle)
		fmt.Println("ConnIdleTime :", ti.ConnIdleTime)
		fmt.Println("TotalTime1    :", rd.Time())
		fmt.Println("TotalTime2    :", ti.TotalTime)
	}
}
