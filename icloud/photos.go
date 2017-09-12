// Copyright 2017 The Xuefei Chen Authors. All rights reserved.
// Created on 2017/9/8 11:52
// Email chenxuefei_pp@163.com

package icloud

import (
    "io/ioutil"
    "encoding/json"
    "net/http"
    "time"
    "encoding/base64"
    "github.com/chenxuefei-pp/ccloudsync/utils"
)


type PhotoAlbum struct {
    RecordName  string
    AlbumName   string
    AlbumType   int64
    Created     time.Time
    Modified    time.Time
}

func (a *PhotoAlbum)  UnmarshalJSON(data []byte) error {
    raw := make(map[string]interface{})
    json.Unmarshal(data,&raw)

    _ ,ok := raw["fields"]
    if !ok {
        return MakeError(BodyFormatError)
    }
    fields := raw["fields"].(map[string]interface{})

    recordName := raw["recordName"].(string)
    albumType := fields["albumType"].(map[string]interface{})["value"].(float64)

    _ ,ok = fields["albumNameEnc"]
    albumName := ""
    if !ok {
        albumName = "Root"
    } else {
        albumName = fields["albumNameEnc"].(map[string]interface{})["value"].(string)
        albumName_bs, _ := base64.StdEncoding.DecodeString(albumName)
        albumName = string(albumName_bs)
    }
    created_ts := int64( raw["created"].(map[string]interface{})["timestamp"].(float64))

    created := time.Unix(created_ts/1e3,(created_ts%1e3)*1e6)
    modified_ts := int64( raw["modified"].(map[string]interface{})["timestamp"].(float64) )
    modified := time.Unix(modified_ts/1e3,(modified_ts%1e3)*1e6)

    a.Created = created
    a.AlbumName = albumName
    a.AlbumType = int64(albumType)
    a.Modified = modified
    a.RecordName = recordName
    return nil
}

type Photo struct {
    bad             bool
    FileName        string
    FileType        string
    FileSize        int64
    DownloadUrl     string
    Width,Height    int
}

func (p *Photo)  UnmarshalJSON(data [] byte) error {
    raw := make(map[string]interface{})
    json.Unmarshal(data,&raw)

    _ ,ok := raw["fields"]
    if !ok {
        return MakeError(BodyFormatError)
    }
    fields := raw["fields"].(map[string]interface{})
    recordType := raw["recordType"].(string)

    if recordType != "CPLMaster"{
        p.bad = true
        return nil
    }

    fileName := fields["filenameEnc"].(map[string]interface{})["value"].(string)
    fileType := fields["itemType"].(map[string]interface{})["value"].(string)
    res := fields["resOriginalRes"].(map[string]interface{})["value"].(map[string]interface{})

    fileSize := res["size"].(float64)
    downloadURL := res["downloadURL"].(string)

    width := fields["resOriginalWidth"].(map[string]interface{})["value"].(float64)
    height := fields["resOriginalHeight"].(map[string]interface{})["value"].(float64)

    fileName_bs, _ := base64.StdEncoding.DecodeString(fileName)
    p.FileName = string(fileName_bs)
    p.FileType = fileType
    p.FileSize = int64(fileSize)
    p.DownloadUrl = downloadURL
    p.Width = int(width)
    p.Height = int(height)
    p.bad = false

    return nil
}



type ICloudPhotos struct {
    srv * ICloudService
    wsUrl string
}

func (p *ICloudPhotos) HandleBodyError(body []byte) error {
    b := make(map[string]interface{})
    json.Unmarshal(body,&b)
    _, ok := b["serverErrorCode"]
    if ok {
        reason := b["reason"].(string)
        utils.Warn("Query Failed : %s", reason)
        return MakeError(QueryFailed)
    }
    return nil
}


func (p *ICloudPhotos) GetToken() error {
    srv := p.srv
    //ckdevice_url := srv.wsUrls["ckdeviceservice"].(map[string]interface{})["url"]
    wsurl := srv.wsUrls["ckdatabasews"].(map[string]interface{})["url"]
    pushws_url := srv.wsUrls["push"].(map[string]interface{})["url"].(string)
    utils.Debug(wsurl.(string))

    params := map[string]string {
        "attempt": "1",
        "clientBuildNumber": client_buildnumber,
        "clientId": srv.clientId,
        "clientMasteringNumber": client_buildnumber,
        "dsid": srv.dsId,
    }

    body := map[string]interface{} {
        "pushTokenTTL": 43200,
        "pushTopics": []string {
            "73f7bfc9253abaaa423eba9a48e9f187994b7bd9",
            "dce593a0ac013016a778712b850dc2cf21af8266",
            "f68850316c5241d8fd120f3bc6da2ff4a6cca9a8",
            "e850b097b840ef10ce5a7ed95b171058c42cc435",
            "8a40cb6b1d3fcd0f5c204504eb8fb9aa64b78faf",
            "5a5fc3a1fea1dfe3770aab71bc46d0aa8a4dad41",
        },
    }

    resp, err := srv.session.Request("POST",pushws_url+"/getToken",params,body)
    if err != nil {
        utils.Warn("error:", err)
        return nil
    }

    headers,_ := json.Marshal(resp.Header)
    resp_body, _ := ioutil.ReadAll(resp.Body)
    //
    defer resp.Body.Close()
    //
    utils.Debug( string( headers ) )

    err = p.HandleBodyError(resp_body)
    if err == nil {
        utils.Debug(string(resp_body))
    } else {
        return err
    }

    return nil
}

func (p *ICloudPhotos) query(payload map[string]interface{}) ( *http.Response, error ) {

    wsurl := p.srv.wsUrls["ckdatabasews"].(map[string]interface{})["url"].(string)
    params := map[string]string {
        "remapEnums": "true",
        "getCurrentSyncToken": "true",
        "clientBuildNumber": client_buildnumber,
        "clientId": p.srv.clientId,
        "clientMasteringNumber": client_buildnumber,
        "dsid": p.srv.dsId,
    }

    resp,err := p.srv.session.Request("POST",wsurl+"/database/1/com.apple.photos.cloud/production/private/records/query",params,payload)
    if err!= nil{
        utils.Warn("error:", err)
        return nil,err
    }
    return resp,nil
}


func (p *ICloudPhotos) QueryList() error {
    srv := p.srv
    wsurl := srv.wsUrls["ckdatabasews"].(map[string]interface{})["url"].(string)
    params := map[string]string {
        "remapEnums": "true",
        "getCurrentSyncToken": "true",
        "clientBuildNumber": client_buildnumber,
        "clientId": srv.clientId,
        "clientMasteringNumber": client_buildnumber,
        "dsid": srv.dsId,
    }

    resp,err := srv.session.Request("GET",wsurl+"/database/1/com.apple.photos.cloud/production/private/zones/list",params,nil)
    if err!= nil{
        utils.Warn("error:", err)
        return nil
    }

    headers,_ := json.Marshal(resp.Header)
    resp_body, _ := ioutil.ReadAll(resp.Body)
    //
    defer resp.Body.Close()
    //
    utils.Debug( string( headers ) )

    err = p.HandleBodyError(resp_body)
    if err == nil {
        utils.Debug(string(resp_body))
    } else {
        return err
    }

    return nil
}

func (p *ICloudPhotos) QueryLookup() error {
    srv := p.srv
    wsurl := srv.wsUrls["ckdatabasews"].(map[string]interface{})["url"].(string)
    params := map[string]string {
        "remapEnums": "true",
        "getCurrentSyncToken": "true",
        "clientBuildNumber": client_buildnumber,
        "clientId": srv.clientId,
        "clientMasteringNumber": client_buildnumber,
        "dsid": srv.dsId,
    }
    body := map[string]interface{} {
        "records": []map[string]string{
            {
                "recordName": "PrimarySync-0000-ZS",
            },
        },
        "zoneID": map[string]string{
            "zoneName": "PrimarySync",
        },
    }

    resp,err := srv.session.Request("POST",wsurl+"/database/1/com.apple.photos.cloud/production/private/records/lookup",params,body)
    if err!= nil{
        utils.Warn("error:", err)
        return nil
    }

    headers,_ := json.Marshal(resp.Header)
    resp_body, _ := ioutil.ReadAll(resp.Body)
    //
    defer resp.Body.Close()
    //
    utils.Debug( string( headers ) )

    err = p.HandleBodyError(resp_body)
    if err == nil {
        utils.Debug(string(resp_body))
    } else {
        return err
    }

    return nil
}

func (p *ICloudPhotos) ListAlbums() ([]PhotoAlbum, error) {
    body := map[string]interface{} {
        "query": map[string]string {
            "recordType": "CPLAlbumByPositionLive",
        },
        "zoneID": map[string]string {
            "zoneName": "PrimarySync",
        },
    }
    resp,err := p.query(body)

    if err!= nil{
        utils.Warn("error:", err)
        return nil,err
    }

    headers,_ := json.Marshal(resp.Header)
    resp_body, _ := ioutil.ReadAll(resp.Body)
    //
    defer resp.Body.Close()
    //
    utils.Debug( string( headers ) )

    err = p.HandleBodyError(resp_body)

    albums_st := struct {
        Records []PhotoAlbum `json:"records"`
        SyncToken string `json:"syncToken"`
    }{}

    if err == nil {
        utils.Debug(string(resp_body))
        err = json.Unmarshal(resp_body,&albums_st)
        if err != nil {
            utils.Error(err.Error())
        }
    } else {
        return nil,err
    }
    return albums_st.Records ,nil
}

func (p *ICloudPhotos) ListAlbumPhotos(album PhotoAlbum) ([]Photo, error) {
    body := map[string]interface{} {
        "desiredKeys" : c_DesiredKeys,
        "query": map[string]interface{} {
            "filterBy": []map[string]interface{}{
                {
                    "comparator": "EQUALS",
                    "fieldName": "parentId",
                    "fieldValue": map[string]string {
                      "value": album.RecordName,
                      "type": "STRING",
                    },
                },
            },
            "recordType": "CPLContainerRelationLiveByAssetDate",
        },
        "zoneID": map[string]string {
            "zoneName": "PrimarySync",
        },
    }
    resp,err := p.query(body)

    if err!= nil{
        utils.Warn("error:", err)
        return nil,err
    }

    headers,_ := json.Marshal(resp.Header)
    resp_body, _ := ioutil.ReadAll(resp.Body)
    //
    defer resp.Body.Close()
    //
    utils.Debug( string( headers ) )

    photos_st := struct {
        Records []Photo `json:"records"`
        SyncToken string `json:"syncToken"`
    }{}

    err = p.HandleBodyError(resp_body)

    if err == nil {
        utils.Debug(string(resp_body))
        err = json.Unmarshal(resp_body,&photos_st)
        if err != nil {
            utils.Error(err.Error())
        }
    } else {
        return nil,err
    }

    results := make([]Photo,0)
    for _,v := range photos_st.Records{
        if !v.bad{
            results = append(results, v)
        }
    }
    return results ,nil
}

func NewPhotoService(srv *ICloudService) *ICloudPhotos {
    return &ICloudPhotos{
        srv: srv,
    }
}
