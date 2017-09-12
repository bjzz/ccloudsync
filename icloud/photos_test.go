// Copyright 2017 The Xuefei Chen Authors. All rights reserved.
// Created on 2017/9/8 16:12
// Email chenxuefei_pp@163.com

package icloud

import (
    "testing"
    "github.com/chenxuefei-pp/ccloudsync/utils"
)

func TestNewPhotoService(t *testing.T) {

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
    p := NewPhotoService(s)

    albums ,_  := p.ListAlbums()

    for _,v := range albums{
        utils.Debug("album %s, created at %s  modified at %s", v.AlbumName,v.Created.String(), v.Modified.String())
    }

    ps , _:= p.ListAlbumPhotos(albums[1])

    for _,v := range ps {
        utils.Debug("PhotoName %s PhotoSize %d PhotoWidth %d PhotoHeight %d",v.FileName,v.FileSize,v.Width, v.Height)
    }
}

func TestICloudSession_Request(t *testing.T) {

}