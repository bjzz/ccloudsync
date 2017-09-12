// Copyright 2017 The Xuefei Chen Authors. All rights reserved.
// Created on 2017/9/6 11:22
// Email chenxuefei_pp@163.com

package icloud

import (
    "testing"
    "net/http"
    "io/ioutil"
    "os"
    "encoding/json"
    "fmt"
    "github.com/chenxuefei-pp/ccloudsync/utils"
)

func Test_CA(t *testing.T){
    resp , err := http.Get(cacert_file_url)
    if err!= nil{
        panic(err.Error())
    }
    cabytes ,err := ioutil.ReadAll(resp.Body)
    file,err := os.Create("F:/Workplace/ca.cert")
    file.Write(cabytes)
}

func TestICloudService_New(t *testing.T) {
    var s * ICloudService
    utils.Debug(s.clientId)
}

func TestICloudService_Auth(t *testing.T) {

    config , err:= utils.ReadJsonConfig("F:\\Workplace\\icloud.json")
    if err != nil {
        t.Fatalf("Read config error ! %s",err)
    }

    var s * ICloudService
    s = s.New(config.Username,config.Password,"F:/Workplace/cookies.sqlite3")
    err = s.Auth()
    if err != nil {
        utils.Warn(err.Error())
    }
}

func TestICloudService_ListTDevice(t *testing.T) {
    config , err:= utils.ReadJsonConfig("F:\\Workplace\\icloud.json")
    if err != nil {
        t.Fatalf("Read config error ! %s",err)
    }
    var s * ICloudService
    s = s.New(config.Username,config.Password,"F:/Workplace/cookies.sqlite3")
    s.Auth()
    devices,_ := s.ListTDevice()

    map_data := devices.(map[string]interface{})

    _,ok := map_data["error"]
    if ok {
        map_data_str ,_:= json.Marshal(map_data)
        utils.Debug( string(map_data_str) )
        utils.Error("List devices failed !")
    }
    arr_data := map_data["devices"].([]interface{})
    s.FetchVerifyCode(arr_data[0])

    var code string
    fmt.Println("Please input your Code: ")
    fmt.Scanln(&code)
    s.SendVerifyCode(arr_data,code)
}