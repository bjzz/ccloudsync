// Copyright 2017 The Xuefei Chen Authors. All rights reserved.
// Created on 2017/9/8 12:54
// Email chenxuefei_pp@163.com

package cookie

import (
    "testing"
    "github.com/chenxuefei-pp/ccloudsync/utils"
)

func TestNewAesEncryptor(t *testing.T) {
    e := NewAesEncryptor("23123asdsdzxcasdqwe")
    edata := e.encode("HellWorld")
    utils.Debug(edata)
    ddata := e.decode(edata)
    utils.Debug(ddata)
}

func TestNewAesEncryptor2(t *testing.T) {

}