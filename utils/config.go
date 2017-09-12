// Copyright 2017 The Xuefei Chen Authors. All rights reserved.
// Created on 2017/9/12 15:32
// Email chenxuefei_pp@163.com

package utils

import (
    "os"
    "io/ioutil"
    "encoding/json"
)

type JsonConfig struct {
    Username    string      `json:"username"`
    Password    string      `json:"password"`
}


func ReadJsonConfig(filename string ) (*JsonConfig, error) {
    f , err := os.Open("F:\\Workplace\\icloud.json")
    if err != nil {
        Warn("Read config file error !")
        return nil,err
    }
    config_bytes , _:= ioutil.ReadAll(f)
    config := JsonConfig{}
    err = json.Unmarshal(config_bytes,&config)
    if err != nil {
        return nil,err
    }
    return &config, nil
}