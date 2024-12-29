package utils

import (
	"fmt"
	"os"
	"strings"
	"time"
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
