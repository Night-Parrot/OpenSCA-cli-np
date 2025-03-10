package walk

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/jlaffaye/ftp"
	"github.com/Night-Parrot/OpenSCA-cli-np/v3/opensca/common"
	"github.com/Night-Parrot/OpenSCA-cli-np/v3/opensca/logs"
)

// isHttp 是否为http/https协议
func isHttp(url string) bool {
	return strings.HasPrefix(url, "http://") ||
		strings.HasPrefix(url, "https://")
}

// isFtp 是否为ftp协议
func isFtp(url string) bool {
	return strings.HasPrefix(url, "ftp://")
}

// isFile 是否为file协议
func isFile(url string) bool {
	return strings.HasPrefix(url, "file://")
}

// download 下载数据
// origin: 数据源
// output: 文件下载路径
// delete: 需要删除的临时文件或目录路径 为空代表不需要删除
func download(origin string) (delete string, output string, err error) {
	defer func() {
		output = filepath.FromSlash(output)
	}()
	if isHttp(origin) {
		tempDir := common.MkdirTemp("download")
		delete = tempDir
		output = filepath.Join(tempDir, filepath.Base(origin))
		err = downloadFromHttp(origin, output)
	} else if isFtp(origin) {
		tempDir := common.MkdirTemp("download")
		delete = tempDir
		output = filepath.Join(tempDir, filepath.Base(origin))
		err = downloadFromFtp(origin, output)
	} else if isFile(origin) {
		output = strings.TrimPrefix(origin, "file:///")
	} else {
		output = origin
	}
	return
}

// downloadFromHttp 下载url并保存到目标文件 支持分片下载
func downloadFromHttp(url, output string) error {

	// 获取head
	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return err
	}
	resp, err := common.HttpDownloadClient.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode > 299 {
		return fmt.Errorf("response code:%d url:%s", resp.StatusCode, url)
	}

	// 创建目标文件
	f, err := os.Create(output)
	if err != nil {
		return err
	}
	defer f.Close()

	// 检测是否支持Accept-Ranges
	if resp.Header.Get("Accept-Ranges") != "bytes" {
		// 不支持分片则尝试直接下载
		r, err := http.Get(url)
		if err != nil {
			return err
		} else {
			defer r.Body.Close()
			size, err := io.Copy(f, r.Body)
			logs.Infof("download %s size:%d", url, size)
			return err
		}
	}

	// 文件总大小
	size, err := strconv.Atoi(resp.Header.Get("Content-Length"))
	if err != nil {
		return err
	}
	offset := 0
	// 分片大小10M
	buffer := 10 * 1024 * 1024

	for offset < size {
		r, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return err
		}
		next := offset + buffer
		if next >= size {
			next = size - 1
		}
		r.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", offset, next))
		resp, err := common.HttpDownloadClient.Do(r)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		_, err = io.Copy(f, resp.Body)
		if err != nil {
			return err
		}
		logs.Infof("download %s range:%d-%d", url, offset, next)
		offset = next + 1
	}
	return nil
}

// downloadFromFtp 下载url并保存到目标文件
func downloadFromFtp(url, output string) error {
	// 解析参数
	var host, path, username, password string
	host = strings.TrimPrefix(url, "ftp://")
	i := strings.Index(host, "/")
	host, path = host[:i], host[i+1:]
	if i := strings.Index(host, "@"); i != -1 {
		up := strings.Split(host[:i], ":")
		if len(up) == 2 {
			username, password = up[0], up[1]
		} else {
			username = host[:i]
		}
		host = host[i+1:]
	}
	if username == "" {
		username = "anonymous"
	}
	// 连接ftp
	c, err := ftp.Dial(host, ftp.DialWithTimeout(5*time.Second))
	if err != nil {
		return err
	}
	defer func() {
		if err := c.Quit(); err != nil {
			fmt.Println(err)
		}
	}()
	// 登录
	err = c.Login(username, password)
	if err != nil {
		return err
	}
	// 获取数据
	r, err := c.Retr(path)
	if err != nil {
		return err
	}
	defer r.Close()
	// 创建目标文件
	f, err := os.Create(output)
	if err != nil {
		return err
	}
	_, err = io.Copy(f, r)
	return err
}
