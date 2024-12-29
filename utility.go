package utils

import (
    "errors"
    "fmt"
    "os"
    "strings"
    "time"

    `github.com/tidwall/gjson`
)

func Join(strs ...string) string {
    var sb = &strings.Builder{}
    defer func() {
        sb = nil
    }()

    for _, str := range strs {
        sb.WriteString(str)
    }
    return sb.String()
}
func FileExists(filename string) bool {
    info, err := os.Stat(filename)
    if os.IsNotExist(err) {
        return false
    }
    return !info.IsDir()
}
func P(s string) {
    fmt.Printf("%s %s \n", gettime(), s)
}
func gettime() string {
    val := time.Now()
    now := val.Format("2006-01-02 15:04:05.000")
    return now
}

func DirectoryExists(filename string) bool {
    info, err := os.Stat(filename)
    if os.IsNotExist(err) {
        return false
    }
    return info.IsDir()
}

func CompressJsonGzip(s *SJson, data string, time2 time.Time) (string, error) {
    var gzip Gzip_Compress
    Fakejson := s.MakeJson(data, time2)
    Cdata, err := gzip.Compress(Fakejson)
    if err != nil {
        return "", err
    }
    return Cdata, nil
}
func DecompressJsonGzip(s *SJson, data string, time2 time.Time) (string, error) {
    var gzip Gzip_Compress
    Cdata, err := gzip.Decompress(data)
    if err != nil {
        return "", err
    }
    Getjson := gjson.Get(Cdata, "data")
    if Getjson.Exists() {
        return Getjson.String(), nil
    } else {
        return "", errors.New("Not Found")
    }
}

func CompressJsonGzipCheck(s *SJson, data string, time2 time.Time) bool {
    var gzip Gzip_Compress
    Cdata, err := gzip.Decompress(data)
    if err != nil {
        return false
    }
    return s.CheckJson(Cdata, time2)
}
