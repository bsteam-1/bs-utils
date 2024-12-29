package utils

import (
    "bytes"
    "encoding/base64"
    "encoding/json"
    "fmt"
    "math/rand"
    "strings"
    "time"

    "github.com/mervick/aes-everywhere/go/aes256"
    "github.com/tidwall/gjson"
)

type SJson struct {
    FakeJsonCount int
    CheckTime     time.Duration
    FakeJsonKey   string
}

func (s *SJson) AESCompressJson(val string, time2 time.Time, pass string) (d string) {
    defer func() {
        s := recover()
        if s != nil {
            d = ""
        }
    }()
    var gzip Gzip_Compress
    Cdata, err := gzip.Compress_B(val)
    if err != nil {
        return ""
    }
    Fakejson := s.MakeJson(base64.StdEncoding.EncodeToString(Cdata), time2)
    d = aes256.Encrypt(Fakejson, pass)
    return d
}

func (s *SJson) AESDecompressJson(val []byte, time2 time.Time, pass string) (d string) {
    defer func() {
        s := recover()
        if s != nil {
            d = ""
        }
    }()
    Dec := aes256.Decrypt(string(val), pass)
    var gzip Gzip_Compress
    if s.CheckJson(Dec, time2) {
        Getjson := gjson.Get(Dec, "data")
        if Getjson.Exists() {
            basedata, _ := base64.StdEncoding.DecodeString(Getjson.String())
            Cdata, err := gzip.Decompress(string(basedata))
            if err != nil {
                return ""
            }
            return Cdata
        } else {
            return ""
        }
    } else {
        return ""
    }
}

func (s *SJson) CompressJson(val string, time2 time.Time) []byte {
    var gzip Gzip_Compress
    Fakejson := s.MakeJson(val, time2)
    Cdata, err := gzip.Compress_B(Fakejson)
    if err != nil {
        return nil
    }
    return Cdata
}

func (s *SJson) DecompressJson(val []byte, time2 time.Time) []byte {
    var gzip Gzip_Compress
    Cdata, err := gzip.Decompress_B(val)
    if err != nil {
        return nil
    }
    Getjson := gjson.Get(string(Cdata), "data")
    if Getjson.Exists() {
        if s.CheckJson(Getjson.String(), time2) {
            return []byte(Getjson.String())
        }
        return nil
    } else {
        return nil
    }
}

func (s *SJson) MakeJson(val string, time2 time.Time) string {
    junkdata := s.makeRandomMap()
    junkdata["time"] = time2.UTC().Format("2006:01:02 15:04:05.000")
    junkdata["data"] = val
    jsondata, _ := json.Marshal(junkdata)
    d6 := s.jsonShuffleObject6(jsondata)
    return d6
}
func (s *SJson) CheckJson(val string, time2 time.Time) bool {
    if bytes.ContainsRune([]byte(val), '{') || bytes.ContainsRune([]byte(val), '}') {
        times, err := time.Parse("2006:01:02 15:04:05.000", gjson.Get(val, "time").String())
        if err != nil {
            return false
        } else {
            if s.inTimeSpan(times, time2) {
                return true
            }
            return false
        }
    } else {
        return false
    }
}

func (s *SJson) inTimeSpan(check time.Time, time2 time.Time) bool {
    local := time2.UTC()
    start := local.Add(-s.CheckTime)
    end := local.Add(s.CheckTime)
    if start.Before(end) {
        return !check.Before(start) && !check.After(end)
    }
    if start.Equal(end) {
        return check.Equal(start)
    }
    return !start.After(check) || !end.Before(check)
}
func (s *SJson) makeRandomMap() map[string]interface{} {
    result := map[string]interface{}{}
    rng := NewUniqueRand(2 * s.FakeJsonCount)
    rng2 := NewUniqueRand(2 * 1000000)
    for i := 0; i < s.FakeJsonCount; i++ {
        result[fmt.Sprintf("%s%d", s.FakeJsonKey, rng.Int())] = rng2.Int() ^ i
    }
    return result
}

func (s *SJson) jsonShuffleObject6(data []byte) string {
    if bytes.ContainsRune(data, '{') || bytes.ContainsRune(data, '}') {
        sb := strings.Builder{}
        sb.WriteString("{")
        rune1 := string(data)
        rune1 = rune1[1 : len(rune1)-1]
        splits := strings.Split(string(rune1), ",")
        rand.Shuffle(len(splits), func(i, j int) {
            splits[i], splits[j] = splits[j], splits[i]
        })
        sb.WriteString(strings.Join(splits, ","))
        sb.WriteString("}")
        return sb.String()
    } else {
        return ""
    }
}

type UniqueRand struct {
    generated map[int]bool // keeps track of
    rng       *rand.Rand   // underlying random number generator
    scope     int          // scope of number to be generated
}

func NewUniqueRand(N int) *UniqueRand {
    s1 := rand.NewSource(time.Now().UnixNano())
    r1 := rand.New(s1)
    return &UniqueRand{
        generated: map[int]bool{},
        rng:       r1,
        scope:     N,
    }
}

func (u *UniqueRand) Int() int {
    if u.scope > 0 && len(u.generated) >= u.scope {
        return -1
    }
    for {
        var i int
        if u.scope > 0 {
            i = u.rng.Int() % u.scope
        } else {
            i = u.rng.Int()
        }
        if !u.generated[i] {
            u.generated[i] = true
            return i
        }
    }
}
