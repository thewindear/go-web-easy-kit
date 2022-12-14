package oauth2

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	timeout               = 10
	githubRedirectBaseUri = "https://github.com/login/oauth/authorize"
	githubAccessToken     = "https://github.com/login/oauth/access_token"
	githubUser            = "https://api.github.com/user"
	githubUsers           = "https://api.github.com/users"
)

type OAuthGithub struct {
	ClientId     string
	ClientSecret string
}

// RedirectUri 生成github授权url跳转地址
func (o *OAuthGithub) RedirectUri(callbackUri, state string) string {
	param := url.Values{}
	param.Add("client_id", o.ClientId)
	param.Add("redirect_uri", callbackUri)
	param.Add("scope", "read:user user:email")
	param.Add("state", state)
	param.Add("allow_signup", "true")
	return githubRedirectBaseUri + "?" + param.Encode()
}

// Code2AccessToken 通过授权的code返回对应 AccessAccessToken
func (o *OAuthGithub) Code2AccessToken(context context.Context, code string) (accessToken *AccessToken, err error) {
	param := url.Values{}
	param.Add("client_id", o.ClientId)
	param.Add("client_secret", o.ClientSecret)
	param.Add("code", code)
	api := githubAccessToken + "?" + param.Encode()
	req, _ := http.NewRequest(http.MethodGet, api, nil)
	result, err := o.request(req, context)
	if err != nil {
		return
	}
	accessToken = &AccessToken{
		Token: result["access_token"].(string),
	}
	return
}

// AccessToken2UserInfo 通过accessToken获取用户信息
func (o *OAuthGithub) AccessToken2UserInfo(context context.Context, accessToken string) (userinfo *UserInfo, err error) {
	req, _ := http.NewRequest(http.MethodGet, githubUser, nil)
	req.Header.Set("Authorization", "token "+accessToken)
	result, err := o.request(req, context)
	if err != nil {
		return
	}
	userinfo = &UserInfo{
		Username: result["login"].(string),
		FirstId:  fmt.Sprintf("%.0f", result["id"]),
		SecondId: result["node_id"].(string),
		Nickname: result["name"].(string),
		Avatar:   result["avatar_url"].(string),
		HomePage: result["html_url"].(string),
		From:     IOAuth2Github,
		Origin:   result,
	}
	if email, ok := result["email"]; ok && email != nil {
		userinfo.Email = email.(string)
	}
	return userinfo, nil
}

// Username2Userinfo github通过使用username获取用户信息
func (o *OAuthGithub) Username2Userinfo(context context.Context, username string) (userinfo *UserInfo, err error) {
	req, _ := http.NewRequest(http.MethodGet, githubUsers+"/"+username, nil)
	result, err := o.request(req, context)
	if err != nil {
		return
	}
	userinfo = &UserInfo{
		Username: result["login"].(string),
		FirstId:  fmt.Sprintf("%.0f", result["id"]),
		SecondId: result["node_id"].(string),
		Nickname: result["name"].(string),
		Avatar:   result["avatar_url"].(string),
		HomePage: result["html_url"].(string),
		Origin:   result,
	}
	return userinfo, nil
}

func (o *OAuthGithub) request(req *http.Request, context context.Context) (map[string]interface{}, error) {
	client := http.DefaultClient
	client.Timeout = timeout * time.Second
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := client.Do(req.WithContext(context))
	if err != nil {
		return nil, fmt.Errorf("oauth2: github request error %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("oauth2: github request [%s] failure http code: %d", req.URL.String(), resp.StatusCode)
		return nil, err
	}
	body, _ := io.ReadAll(resp.Body)
	var tmp map[string]interface{}
	_ = json.Unmarshal(body, &tmp)
	if _, ok := tmp["error"]; ok {
		err = fmt.Errorf("oauth2: github request [%s] failure error: %s reason: %s", req.URL.String(), tmp["error"], tmp["error_description"])
		return nil, err
	}
	return tmp, nil
}
