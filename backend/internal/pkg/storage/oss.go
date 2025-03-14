package storage

import (
	"bytes"
	"fmt"
	"lottery-backend/internal/config"
	"path/filepath"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

var ossClient *oss.Client
var bucket *oss.Bucket

// InitOSS 初始化OSS客户端
func InitOSS() error {
	var err error
	ossClient, err = oss.New(
		config.Current.AliyunOSS.Endpoint,
		config.Current.AliyunOSS.AccessKeyID,
		config.Current.AliyunOSS.AccessKeySecret,
	)
	if err != nil {
		return fmt.Errorf("初始化OSS客户端失败: %v", err)
	}

	bucket, err = ossClient.Bucket(config.Current.AliyunOSS.BucketName)
	if err != nil {
		return fmt.Errorf("获取OSS Bucket失败: %v", err)
	}

	return nil
}

// UploadTicketImage 上传彩票图片到OSS
func UploadTicketImage(data []byte, fileName string) (string, error) {
	// 生成OSS中的文件路径
	now := time.Now()
	objectKey := fmt.Sprintf("tickets/%d/%02d/%02d/%s",
		now.Year(), now.Month(), now.Day(),
		filepath.Base(fileName))

	// 上传文件
	err := bucket.PutObject(objectKey, bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("上传文件到OSS失败: %v", err)
	}

	// 返回可访问的URL
	return fmt.Sprintf("%s/%s", config.Current.AliyunOSS.BucketDomain, objectKey), nil
}

// DeleteTicketImage 从OSS删除彩票图片
func DeleteTicketImage(imageUrl string) error {
	// 从URL中提取objectKey
	objectKey := filepath.Base(imageUrl)
	err := bucket.DeleteObject(objectKey)
	if err != nil {
		return fmt.Errorf("从OSS删除文件失败: %v", err)
	}
	return nil
}
