// Copyright 2017 The Xuefei Chen Authors. All rights reserved.
// Created on 2017/9/5 17:29
// Email chenxuefei_pp@163.com

package utils

import "testing"

func TestLogger_SetConfig(t *testing.T) {

    config := Logger.GetConfig()
    config.FileName = "F:/Workplace/1.log"
    Logger.UpdateConfig(config)
}