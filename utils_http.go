package cvvapi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

var httpPipeline HttpHandler = doRequest

func doRequest(req *http.Request) (*http.Response, error) {
	resp, err := config.GetClient().Do(req)

	if err != nil {
		return nil, NewNetworkError(err)
	}

	return resp, nil
}

type httpResponse http.Response

type HttpHandler func (*http.Request) (*http.Response, error)
type HttpMiddleware func (*http.Request, HttpHandler) (*http.Response, error)

func setSessionCookie(user *Session, request *http.Request) {
	request.Header.Add("Cookie", user.sessionCookiesHeader)
}

func checkSuccessStatusCode(response *http.Response) error {
	if response.StatusCode >= 200 && response.StatusCode < 400 {
		return nil
	}

	return NewHttpError(fmt.Errorf("checkSuccessStatusCode"), response.StatusCode)
}

func (o *Session) doRequest(request *http.Request) (*http.Response, error) {
	setSessionCookie(o, request)

	var resp *http.Response
	resp, err := httpPipeline(request)

	if err != nil {
		return nil, err
	}

	err = checkSuccessStatusCode(resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (o *Session) doGet(requestUrl string) (*httpResponse, error) {
	var request, err = http.NewRequest("GET", requestUrl, nil)
	if err != nil {
		return nil, err
	}

	resp, err := o.doRequest(request)
	
	return (*httpResponse)(resp), err
}

func (o *Session) doPost(requestUrl string, body io.Reader) (*httpResponse, error) {
	var request, err = http.NewRequest("POST", requestUrl, body)
	if err != nil {
		return nil, err
	}

	resp, err := o.doRequest(request)
	
	return (*httpResponse)(resp), err
}

func (o *Session) doPostForm(requestUrl string, values url.Values)(*httpResponse, error) {
	var request, err = http.NewRequest("POST", requestUrl, strings.NewReader(values.Encode()))
	if err != nil {
		return nil, err
	}
	
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := o.doRequest(request)
	
	return (*httpResponse)(resp), err
}

func (o *Session) doPostJson(requestUrl string, jsonText string)(*httpResponse, error) {
	var request, err = http.NewRequest("POST", requestUrl, strings.NewReader(jsonText))
	if err != nil {
		return nil, err
	}

	request.Header.Set("Content-Type", "application/json")

	resp, err := o.doRequest(request)
	
	return (*httpResponse)(resp), err
}

func (o *httpResponse) getText() (string, error) {
	defer o.Body.Close()

	var r, err = io.ReadAll(o.Body)
	if err != nil {
		return "", err
	}

	return string(r), err
}

func (o *httpResponse) getObject(obj any) error {
	defer o.Body.Close()

	var err = json.NewDecoder(o.Body).Decode(obj)

	if err != nil {
		return err
	}

	return nil
}

func (o *httpResponse) getHtmlNode() *htmlNode {
	defer o.Body.Close()

	return parseHtmlReader(o.Body)
}