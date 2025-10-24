package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/qiniu/go-sdk/v7/storagev2/credentials"
	"github.com/qiniu/go-sdk/v7/storagev2/uptoken"
)

type Config struct {
	AccessKey   string `json:"ak"`
	SecretKey   string `json:"sk"`
	Bucket      string `json:"bucket"`
	ExpiresHour int    `json:"expires_hour"`
	Domain      string `json:"domain"`
}

type Storage struct {
	conf Config
}

func NewStorage(conf Config) (*Storage, error) {
	if conf.AccessKey == "" || conf.SecretKey == "" || conf.Bucket == "" {
		return nil, errors.New("invalid ak or sk or bucket")
	}
	if conf.ExpiresHour == 0 {
		conf.ExpiresHour = 2
	}
	return &Storage{
		conf: conf,
	}, nil
}

func (s *Storage) GenerateUploadToken(userID int64) (string, error) {
	saveKey := fmt.Sprintf("voices/${year}/${mon}/${day}/${hour}${min}${sec}-%d-${fname}", userID)
	mac := credentials.NewCredentials(s.conf.AccessKey, s.conf.SecretKey)
	policy, err := uptoken.NewPutPolicy(s.conf.Bucket, time.Now().Add(time.Duration(s.conf.ExpiresHour)*time.Hour))
	if err != nil {
		return "", err
	}
	policy.SetReturnBody(`{"key":"$(key)","hash":"$(etag)","fsize":$(fsize),"type":$(mimeType)}`).
		SetForceSaveKey(true).
		SetSaveKey(saveKey)
	return uptoken.NewSigner(policy, mac).GetUpToken(context.Background())
}

func (s *Storage) MakeURL(key string) string {
	return "https://" + s.conf.Domain + "/" + key
}

type UploadFileRet struct {
	Key    string
	Hash   string
	Fsize  int
	Bucket string
	Name   string
}
