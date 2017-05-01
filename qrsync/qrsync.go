// qrsync.go
package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"qiniupkg.com/api.v7/conf"
	"qiniupkg.com/api.v7/kodo"
	"qiniupkg.com/api.v7/kodocli"
	"strings"
	"sync"
)

var (
	flagDir    = flag.String("dir", "./", "dir [Default: ./]")
	flagAk     = flag.String("ak", "", "qiniu ak (require)")
	flagSk     = flag.String("sk", "", "qiniu sk (require)")
	flagBucket = flag.String("bucket", "", "qiniu bucket (require)[Example: bucket1|bucket2]")
	help       = flag.Bool("h", false, "help")
	version    = flag.Bool("version", false, "qrsync version")
	loger      = log.New(os.Stdout, "", log.LstdFlags)
	wg         sync.WaitGroup
)

func main() {
	loger.Println("rsync start...")
	flag.Parse()
	if *help {
		Help()
		return
	}
	if *version {
		Version()
		return
	}
	if *flagAk == "" || *flagSk == "" {
		loger.Println("qiniu ak or sk err:", "is null")
		return
	}
	conf.ACCESS_KEY = *flagAk
	conf.SECRET_KEY = *flagSk

	if !IsExist(*flagDir) {
		loger.Println("dir is not exist err:", *flagDir)
		return
	}
	err := os.Chdir(*flagDir)
	if err != nil {
		loger.Println("chdir err", err)
		return
	}
	var fileinfos []string
	err = filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		fileinfos = append(fileinfos, path)
		return nil
	})
	if err != nil {
		loger.Println("filepath walk err:", err)
		return
	}
	if len(fileinfos) == 0 {
		loger.Println("dir err:", "no file in the directory")
		return
	}
	buckets := strings.Split(*flagBucket, "|")
	for _, bucket := range buckets {
		wg.Add(1)
		go func(bucket string) {
			for _, v := range fileinfos {
				_, err = QiniuUpload(bucket, v, v)
				if err != nil {
					loger.Println(bucket, "qiniu upload err:", err)
				}
				loger.Println(bucket, "qiniu upload success:", v)
			}
			wg.Done()
		}(bucket)

	}
	wg.Wait()
	loger.Println("qrsync success...")

}
func IsExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}

func QiniuUpload(bucket, dest, src string) (ret PutRet, err error) {
	c := kodo.New(0, nil)
	policy := &kodo.PutPolicy{
		Scope:   bucket + ":" + dest,
		Expires: 3600,
	}
	token := c.MakeUptoken(policy)
	zone := 0
	uploader := kodocli.NewUploader(zone, nil)
	err = uploader.PutFile(nil, &ret, token, dest, src, nil)
	return
}

// 构造返回值字段
type PutRet struct {
	Hash string `json:"hash"`
	Key  string `json:"key"`
}

func Help() {
	loger.Printf(
		"\nUseage: %s [ Options ]\n\n"+
			"Options:\n"+
			"  -dir           dir [Default: ./]\n"+
			"  -ak            qiniu ak (require)\n"+
			"  -sk            qiniu sk (require)\n"+
			"  -bucket        qiniu bucket (require)[Example: bucket1|bucket2]\n"+
			"  -h             help\n"+
			"  -version       qrsync version\n"+
			"------------------------------------------------------\n\n",
		os.Args[0])

	os.Exit(0)
}
func Version() {
	loger.Println("qrsync version:1.0")
}
