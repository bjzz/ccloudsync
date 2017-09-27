// Copyright 2017 The Xuefei Chen Authors. All rights reserved.
// Created on 2017/9/27 11:16
// Email chenxuefei_pp@163.com

package ccloudsync


type CCError interface {
    error
    Equal(error int) bool
}

