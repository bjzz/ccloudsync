// Copyright 2017 The Xuefei Chen Authors. All rights reserved.
// Created on 2017/9/5 15:31
// Email chenxuefei_pp@163.com

package icloud

import (
    "net/http"
    "crypto/x509"
    "io/ioutil"
    "crypto/tls"
    "github.com/satori/go.uuid"
    "strings"
    "encoding/json"
    "bytes"
    "compress/gzip"
    "net/url"
    "os"
    "github.com/chenxuefei-pp/go-cookie"
    "github.com/chenxuefei-pp/ccloudsync/utils"
)

const (
    home_endpoint      = "https://www.icloud.com"
    setup_endpoint     = "https://setup.icloud.com/setup/ws/1"
    client_buildnumber = "17EHotfix"
    cacert_file_url    = "http://curl.haxx.se/ca/cacert.pem"
)

type ICloudSession struct {
    http.Client
    jar     http.CookieJar
    cookies [] *http.Cookie
    headers http.Header
}

func (*ICloudSession) New(Id string, CookiePath string) (*ICloudSession) {
    cookiePath := ""
    if CookiePath == ""{
        cookiePath = os.TempDir() + "/ccloudsync_cookie.sqlite3"
    } else {
        cookiePath = CookiePath
    }

    c := cookie.NewSqliteJar(cookiePath)

    pool := x509.NewCertPool()

    resp, err := http.Get(cacert_file_url)
    if err != nil {
        utils.Error("Read CA %s Failed!", cacert_file_url)
    }

    caCert, _ := ioutil.ReadAll(resp.Body)
    pool.AppendCertsFromPEM(caCert)

    tr := &http.Transport{
        TLSClientConfig: &tls.Config{RootCAs: pool},
    }

    return &ICloudSession{
        jar: c,
        Client: http.Client{
            Transport:     tr,
            CheckRedirect: nil,
            Jar:           c,
        },
        headers: make(http.Header),
    }
}


// HTTP Request 实现
func (s * ICloudSession) Request(method string, _url string, _params map[string]string , _body interface{}) (*http.Response, error) {
    bf := bytes.Buffer{}
    if _body != nil {
        body, _ := json.Marshal(_body)
        bf.Write(body)
    }
    req_url,err := url.Parse(_url)
    if err != nil{
        return nil,MakeError(UrlEncodeError)
    }
    querys := req_url.Query()
    for k,v := range _params{
        querys.Add(k,v)
    }
    req_url.RawQuery = querys.Encode()

    req, err := http.NewRequest(method, req_url.String(),&bf)

    resp, err := s.Do(req)
    if err != nil {
        utils.Warn("error:", err)
        return nil,err
    }
    return resp,nil
}

// 实际Request
func (s *ICloudSession) Do(req *http.Request) (*http.Response, error) {
    for k, v := range s.headers {
        _, ok := req.Header[k]
        if !ok {
            req.Header[k] = v
        }
    }

    resp, err := s.Client.Do(req)

    if err != nil {
        utils.Warn(err.Error())
        return resp, err
    }

    switch resp.Header.Get("Content-Encoding") {
    case "gzip":
        reader, err := gzip.NewReader(resp.Body)
        if err != nil {
            utils.Error("Create gzip reader failed !", err)
        }
        resp.Body = reader
    default:
    }

    s.cookies = s.jar.Cookies(req.URL)

    return resp, nil
}

type ICloudService struct {
    session  *ICloudSession
    appleId  string
    password string
    clientId string
    dsId     string
    wsUrls   map[string]interface{}
}

// 创建新的Service
func (*ICloudService) New(Id string, Pwd string, CookiePath string ) (*ICloudService) {
    var sn *ICloudSession

    s := &ICloudService {
        session:  sn.New(Id, CookiePath),
        appleId:  Id,
        password: Pwd,
        wsUrls:   make(map[string]interface{}),
    }

    s.session.headers.Add("Accept", "*/*")
    s.session.headers.Add("Accept-Encoding", "gzip, deflate, br")
    s.session.headers.Add("Content-Type", "text/plain")
    s.session.headers.Add("Origin", home_endpoint)
    s.session.headers.Add("Referer", home_endpoint+"/")
    s.session.headers.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/60.0.3112.78 Safari/537.36")

    _uid := uuid.NewV4()

    s.clientId = strings.ToUpper(_uid.String())
    return s
}

// 认证
func (s *ICloudService) Auth() *ICloudError {

    body := map[string]interface{}{
        "apple_id": s.appleId,
        "password": s.password,
    }

    params := map[string]string {
        "clientBuildNumber": client_buildnumber,
        "clientId": s.clientId,
        "clientMasteringNumber": client_buildnumber,
    }

    resp , err := s.session.Request("POST",setup_endpoint+"/accountLogin",params,body)

    if err != nil {
        utils.Warn("error:", err)
        return MakeError(RequestError)
    }

    resp_body, _ := ioutil.ReadAll(resp.Body)

    defer resp.Body.Close()

    resp_data := make(map[string]interface{})
    json.Unmarshal(resp_body, &resp_data)

    if resp_data["hsaChallengeRequired"].(bool) {
        return MakeError(ChallengeRequired)
    }

    _, ok := resp_data["webservices"]
    if ok {
        s.wsUrls, ok = resp_data["webservices"].(map[string]interface{})
        if !ok {
            return MakeError(ConversionTypeError)
        }
    } else {
        return MakeError(BodyFormatError)
    }

    _, ok = resp_data["dsInfo"]
    if ok {
        dsInfo, ok := resp_data["dsInfo"].(map[string]interface{})
        if !ok {
            return MakeError(ConversionTypeError)
        }
        s.dsId = dsInfo["dsid"].(string)
    } else {
        return MakeError(BodyFormatError)
    }

    return nil
}

// 列出支持设备
func (s *ICloudService) ListTDevice() (interface{}, error) {
    params := map[string]string {
        "clientBuildNumber": client_buildnumber,
        "clientId": s.clientId,
    }

    resp,err := s.session.Request("POST",setup_endpoint+"/listDevices",params,nil)
    //resp, err := s.session.Do(req)
    if err != nil {
        utils.Warn("error:", err)
        return nil, err
    }

    //headers,_ := json.Marshal(resp.Header)
    resp_body, _ := ioutil.ReadAll(resp.Body)

    defer resp.Body.Close()

    //utils.Debug( string( headers ) )
    //utils.Debug( string( resp_body ) )

    var data interface{}
    err = json.Unmarshal(resp_body, &data)

    if err != nil {
        return nil, err
    }
    return data, nil
}

// 发送二次验证
func (s *ICloudService) FetchVerifyCode(device interface{}) error {

    params := map[string]string {
        "clientBuildNumber": client_buildnumber,
        "clientId": s.clientId,
    }

    resp, err := s.session.Request("POST",setup_endpoint+ "/sendVerificationCode",params, device)

    if err != nil {
        utils.Warn("Request verfify code error : %s",err)
        return err
    }
    //headers,_ := json.Marshal(resp.Header)
    //resp_body,_ := ioutil.ReadAll(resp.Body)
    //
    defer resp.Body.Close()
    //
    //utils.Debug( string( headers ) )
    //utils.Debug( string( resp_body ) )

    return nil
}

func (s *ICloudService) SendVerifyCode(device interface{}, code string) error {
    map_device, ok := device.(map[string]interface{})
    if !ok {
        utils.Warn("Device info error !")
        return MakeError(DeviceInfoError)
    }
    map_device["verificationCode"] = code
    map_device["trustBrowser"] = true

    params := map[string]string{
        "clientBuildNumber": client_buildnumber,
        "clientId": s.clientId,
    }

    resp, err := s.session.Request("POST",setup_endpoint+"/validateVerificationCode",params,map_device)
    defer resp.Body.Close()

    if err != nil {
        utils.Warn("Send verify code error ! %s", err)
        return err
    }
    return nil
}
