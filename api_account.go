package cvvapi

import (
	"fmt"
	"net/url"
	"slices"
	"strings"
	"time"
)

// Session of interaction with ClasseViva. It is able to refresh the session token if necessary.
type Session struct {
	Username             string
	password string
	sessionCookiesHeader string
	age time.Time
}

// Creates a new Session, storing the necessary information to refresh the token.
func NewSession(username string, password string) (*Session, error) {
	
	var session = &Session{Username: username, password: password}
	
	err := session.Refresh()

	if err != nil {
		return nil, err
	}

	return session, nil
}

// Verify if a session is stil usable
func (o *Session) IsActive() bool {

	return o.sessionCookiesHeader != ""  && time.Since(o.age) < config.sessionCheckTimeout
}

// Check if a session is stil usable Asking ClasseViva website
func (o *Session) CheckActive() (bool, error) {
	
	reqUrl := fmt.Sprintf("https://%v/sps/app/default/SocMsgApi.php?a=acGetUnreadCount", config.cvvHostname)

	resp, err := o.doGet(reqUrl)

	if err != nil {
		return false, err
	}

	var data = new(struct {
		Errors []string `json:"error"`
	})
	err = resp.getObject(data)
	
	if err != nil {
		return false, NewApiError(err)
	}

	active := slices.IndexFunc(data.Errors, func (p string) bool { 
		return strings.HasPrefix(p, "001/not authenticated") }) < 0

	if active {
		o.age = time.Now()
	}

	return active, nil
}

// Refresh the session performing a new login to ClasseViva if necessary.
func (o *Session) Refresh() error {
	signInUrl := fmt.Sprintf("https://%v/auth-p7/app/default/AuthApi4.php?a=aLoginPwd", config.cvvHostname)
	
	resp, err := o.doPostForm(signInUrl, url.Values{
		"uid": []string{o.Username},
		"pwd": []string{o.password},
	})
	
	if err != nil {
		return err
	}

	var result = new(struct {
		Data struct {
			Auth struct {
				Verified bool	`json:"verified"`
				LoggedIn bool `json:"loggedIn"`
			} `json:"auth"`
		} `json:"data"`
	})
	err = resp.getObject(result)
	
	if err != nil {
		return NewApiError(err)
	}
	
	if !(result.Data.Auth.Verified && result.Data.Auth.LoggedIn) {

		return NewApiError(fmt.Errorf("unable to authenticate (verified=%v, loggedIn=%v)", 
			result.Data.Auth.Verified, result.Data.Auth.LoggedIn))
	}

	o.sessionCookiesHeader = resp.Header.Values("Set-Cookie")[1]
	o.age = time.Now()
	
	return nil
}

func (o *Session) EnsureActive() error {
	if o.IsActive() {
		return nil
	}

	ok, err := o.CheckActive()
	if err != nil {
		return err
	}

	if !ok {
		err := o.Refresh()

		if err != nil {
			return err
		}
	}

	return nil
}