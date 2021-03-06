package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"

	"github.com/xeipuuv/gojsonschema"
)

var headerJar = make(map[string]string)

type Tests []*Test

type responseHeaderTestCase struct {
	Key      string `json:"key"`
	Value    string `json:"value"`
	Validate bool   `json:"validate"`
	PutInJar bool   `json:"putInJar"`
}

type Test struct {
	Label   string        `json:"label"`
	request *http.Request // contains the actual request
	Request *struct {
		Method      string   `json:"method"`
		URL         *url.URL `json:"url"`
		URLUserInfo *struct {
			User     string `json:"user"`
			Password string `json:"password"`
		} `json:"urlUserInfo"`
		TLSInsecureSkipVerify bool `json:"tlsInsecureSkipverify"`
		NoDefaultHeaders      bool `json:"noDefaultHeaders"`
		Headers               []*struct {
			Key        string `json:"key"`
			Value      string `json:"value"`
			UseFromJar bool   `json:"useFromJar"`
		} `json:"headers"`
		BodyString string      `json:"bodyString"`
		BodyJson   interface{} `json:"bodyJson"`
	} `json:"request"`
	response *http.Response // contains the actual response
	Response *struct {
		Status           string                    `json:"status,omitempty`
		StatusCode       int                       `json:"statusCode"`
		NoDefaultHeaders bool                      `json:"noDefaultHeaders"`
		Headers          []*responseHeaderTestCase `json:"headers"`
		contentType      string
		BodyCheck        bool                   `json:"bodyCheck"`
		BodyString       string                 `json:"bodyString"`
		BodyJsonSchema   map[string]interface{} `json:"bodyJsonSchema"`
		body             []byte
	} `json:"response"`
	UseCookieJar      bool `json:"useCookieJar"`
	NoCookieJar       bool `json:"NoCookieJar"`
	cookieJar         *cookiejar.Jar
	PrintDebugOnFail  bool `json:"printDebugOnFail`
	PrintJsonIndented bool `json:"printJsonIndented"`
	failed            bool
}

func (t *Test) Run() bool {
	defer t.printDebugOnfail()
	if t.failed {
		return false
	}
	var err error
	t.response, err = Call(t.request, t.Request.TLSInsecureSkipVerify)
	if err != nil {
		t.fail(err)
		return false
	}
	t.evaluate()
	return !t.failed
}

func (t *Test) Prepare(defaultTest *Test) {
	if t.Request == nil {
		t.fail(errors.New("cannot execute test because 'request' is missing"))
		return
	}

	if defaultTest != nil {
		if defaultTest.PrintDebugOnFail {
			t.PrintDebugOnFail = true
			if defaultTest.PrintJsonIndented {
				t.PrintJsonIndented = true
			}
		}
	}
	if t.Request.Method == "" && defaultTest != nil {
		if defaultTest.Request.Method == "" {
			t.fail(errors.New("request method missing"))
		} else {
			t.Request.Method = defaultTest.Request.Method
		}
	}
	t.prepareURL(defaultTest)
	var body io.Reader
	var err error
	if t.Request.BodyString != "" {
		body = bytes.NewBufferString(t.Request.BodyString)
	} else if t.Request.BodyJson != nil {
		var b []byte
		b, err = json.Marshal(t.Request.BodyJson)
		if err != nil {
			t.fail(err)
			return
		}
		body = bytes.NewBuffer(b)
	}
	t.request, err = http.NewRequest(t.Request.Method, t.Request.URL.String(), body)
	if err != nil {
		t.fail(err)
	}
	t.prepareHeaders(defaultTest)
	t.prepareCookies(defaultTest)
}

func (t *Test) prepareURL(defaultTest *Test) {
	if t.Request.URL == nil {
		return
	}
	if t.Request.URLUserInfo != nil && t.Request.URLUserInfo.User != "" {
		t.Request.URL.User = url.UserPassword(t.Request.URLUserInfo.User, t.Request.URLUserInfo.Password)
	}
	if defaultTest == nil || defaultTest.Request == nil || defaultTest.Request.URL == nil {
		return
	}
	if t.Request.URL.Scheme == "" {
		t.Request.URL.Scheme = defaultTest.Request.URL.Scheme
	}
	if t.Request.URL.Opaque == "" {
		t.Request.URL.Opaque = defaultTest.Request.URL.Opaque
	}
	if t.Request.URL.Host == "" {
		if t.Request.URL.User == nil {
			t.Request.URL.User = defaultTest.Request.URL.User
		}
		t.Request.URL.Host = defaultTest.Request.URL.Host
		if defaultTest.Request.TLSInsecureSkipVerify {
			t.Request.TLSInsecureSkipVerify = true
		}
	}
	if t.Request.URL.Path == "" {
		t.Request.URL.Path = defaultTest.Request.URL.Path
	}
	if t.Request.URL.RawQuery == "" {
		t.Request.URL.RawQuery = defaultTest.Request.URL.RawQuery
	}
	if t.Request.URL.Fragment == "" {
		t.Request.URL.Fragment = defaultTest.Request.URL.Fragment
	}
	url := t.Request.URL.String()
	if url == "" {
		t.fail(errors.New("request url missing"))
	}
}

func (t *Test) prepareHeaders(defaultTest *Test) {
	// set request headers
	if t.Request.NoDefaultHeaders == false && defaultTest != nil && defaultTest.Request != nil && defaultTest.Request.Headers != nil {
		//TODO test if the default headers get overwritten by the ones in described in the test
		t.Request.Headers = append(defaultTest.Request.Headers, t.Request.Headers...)
	}

	if len(t.Request.Headers) > 0 {
		for _, h := range t.Request.Headers {
			if h.UseFromJar {
				if v, ok := headerJar[h.Key]; ok {
					h.Value = v
				}
			}
			t.request.Header.Add(h.Key, h.Value)
		}
	}

	// set response test case headers
	if t.Response != nil && t.Response.NoDefaultHeaders == false && defaultTest != nil && defaultTest.Response != nil && defaultTest.Response.Headers != nil {
		testCasesToAdd := make([]*responseHeaderTestCase, 0)
		for _, defaultTestCase := range defaultTest.Response.Headers {
			found := false
			for _, testCase := range t.Response.Headers {
				if defaultTestCase.Key == testCase.Key {
					found = true
				}
			}
			if !found {
				testCasesToAdd = append(testCasesToAdd, defaultTestCase)
			}
		}
		t.Response.Headers = append(t.Response.Headers, testCasesToAdd...)
	}
}

func (t *Test) prepareCookies(defaultTest *Test) {
	if defaultTest == nil {
		// t is the default test create new cookie jar in it nothing more
		var err error
		if t.cookieJar, err = cookiejar.New(nil); err != nil {
			t.fail(err)
		}
		return
	}
	// no cookie jar precede above use cookie jar
	if t.NoCookieJar {
		return
	}
	if defaultTest.UseCookieJar {
		t.UseCookieJar = true
	}
	if !t.UseCookieJar {
		return
	}
	t.cookieJar = defaultTest.cookieJar
	for _, c := range t.cookieJar.Cookies(t.Request.URL) {
		t.request.AddCookie(c)
	}
}

func (t *Test) evaluate() {
	t.readResponse()
	if t.Response == nil {
		return
	}
	t.evaluateHeaders()
	t.evaluateStatusCode()
	t.evaluateStatus()
	t.evaluateBody()
}

func (t *Test) readResponse() {
	if t.response == nil {
		return
	}
	// content type
	if t.response.Header != nil {
		if v, ok := t.response.Header["Content-Type"]; ok {
			t.Response.contentType = v[0]
		}
	}
	// cookies
	if cookies := t.response.Cookies(); cookies != nil && !t.NoCookieJar && t.UseCookieJar {
		t.cookieJar.SetCookies(t.Request.URL, cookies)
	}
	// body
	if t.Response.body != nil {
		return
	}
	var err error
	t.Response.body, err = ioutil.ReadAll(t.response.Body)
	defer t.response.Body.Close()
	if err != nil {
		t.fail(fmt.Errorf("response body read error ", err))
	}
}

func (t *Test) evaluateHeaders() {
	if t.Response.Headers != nil {
		for _, testCase := range t.Response.Headers {
			value, ok := t.response.Header[testCase.Key]
			if ok {
				if testCase.PutInJar && value[0] == testCase.Value {
					headerJar[testCase.Key] = value[0]
				} else if testCase.Validate && value[0] != testCase.Value {
					t.fail(fmt.Errorf("expected header %s to equal %s, given %s", testCase.Key, testCase.Value, value[0]))
				}
			} else if testCase.Validate {
				t.fail(fmt.Errorf("expected header %s to be present", testCase.Key))
			}
		}
	}
}

func (t *Test) evaluateStatusCode() {
	if t.Response.StatusCode != 0 && t.Response.StatusCode != t.response.StatusCode {
		t.fail(fmt.Errorf("expect status code to equal %d, given %d", t.Response.StatusCode, t.response.StatusCode))
	}
}

func (t *Test) evaluateStatus() {
	if t.Response.Status != "" && t.Response.Status != t.response.Status {
		t.fail(fmt.Errorf("expect status to equal %q, given %q", t.Response.Status, t.response.Status))
	}
}

func (t *Test) evaluateBody() {
	if !t.Response.BodyCheck {
		return
	}
	if t.Response.BodyString != "" {
		b := []byte(t.Response.BodyString)
		if bytes.Equal(t.Response.body, b) {
			return
		}
		t.fail(fmt.Errorf("expect response body to equal %q, given %q", t.Response.BodyString, t.Response.body))
	} else if t.Response.BodyJsonSchema != nil {
		var v interface{}
		if err := json.Unmarshal(t.Response.body, &v); err != nil {
			t.fail(fmt.Errorf("response json body error %s", err))
			return
		}
		result, err := gojsonschema.Validate(gojsonschema.NewGoLoader(t.Response.BodyJsonSchema), gojsonschema.NewGoLoader(v))
		if err != nil {
			t.fail(fmt.Errorf("validation error %s", err))
			return
		}
		if !result.Valid() {
			for _, desc := range result.Errors() {
				t.fail(fmt.Errorf("JSON schema expect %s", desc))
			}
		}
	}
}

func (t *Test) fail(err error) {
	if t.failed == false {
		t.failed = true
		fmt.Println("\n\033[1;31mFAILED\033[0m", t.Label)
	}
	fmt.Printf("  \033[0;31m%s\033[0m\n", err)
}

func (t *Test) printDebugOnfail() {
	if t.failed && t.PrintDebugOnFail {
		fmt.Println("\033[1;36mDEBUG REQUEST\033[0m")
		// request
		fmt.Printf("  \033[1;33mURL\033[0m: %s\n", t.Request.URL.String())
		fmt.Printf("  \033[1;33mHeaders\033[0m: %+v\n", t.request.Header)
		fmt.Printf("  \033[1;33mBody\033[0m: ")
		if t.Request.BodyString != "" {
			fmt.Println(t.Request.BodyString)
		} else if t.Request.BodyJson != nil {
			b, err := json.Marshal(t.Request.BodyJson)
			if err != nil {
				t.fail(err)
				return
			}
			if t.PrintJsonIndented {
				var out bytes.Buffer
				if err = json.Indent(&out, b, "\t", "  "); err != nil {
					fmt.Println(err)
				} else {
					fmt.Println(out.String())
				}
			} else {
				fmt.Printf("%s\n", b)
			}
		} else {
			fmt.Println("")
		}
		//response
		if t.Response != nil {
			fmt.Println("\033[1;36mDEBUG RESPONSE\033[0m")
			if t.response != nil {
				fmt.Printf("  \033[1;33mHeaders\033[0m: %+v\n", t.response.Header)
				fmt.Printf("  \033[1;33mStatus code\033[0m: %+v\n", t.response.StatusCode)
				fmt.Printf("  \033[1;33mStatus\033[0m: %+v\n", t.response.Status)
				if t.Response.body != nil {
					defaultBodyPrint := false
					fmt.Printf("  \033[1;33mBody\033[0m: ")
					if strings.Contains(strings.ToLower(t.Response.contentType), "application/json") {
						if t.PrintJsonIndented {
							var out bytes.Buffer
							if err := json.Indent(&out, t.Response.body, "\t", "  "); err != nil {
								fmt.Println(err)
							} else {
								fmt.Println(out.String())
							}
						} else {
							defaultBodyPrint = true
						}
					}
					if defaultBodyPrint {
						fmt.Printf("%s\n", t.Response.body)
					}
				}
			} else {
				fmt.Println("  \033[0;31mno response\033[0m")
			}
		}
	}
}
