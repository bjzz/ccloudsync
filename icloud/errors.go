// Copyright 2017 The Xuefei Chen Authors. All rights reserved.
// Created on 2017/9/6 16:21
// Email chenxuefei_pp@163.com

package icloud

const (
    ChallengeRequired   = 1
    ConversionTypeError = 2
    BodyFormatError     = 3
    DeviceInfoError     = 4
    UrlEncodeError      = 5
    QueryFailed         = 6
)

type ICloudError int

func MakeError(code int) *ICloudError {
    e := new(ICloudError)
    *e = ICloudError(code)
    return e
}


type TypeError int

func (t *TypeError) Error() string {
    return "Type passed error!"
}

func (t *ICloudError) Error() string {
    switch *t {
    case ChallengeRequired:
        return "2FA required!"
    case ConversionTypeError:
        return "Convertiosn type error !"
    case BodyFormatError:
        return "Body format error!Cannot convert!"
    default:
        return "Unknow error!"
    }
}
