package tokens

import (
	"regexp"
	"strings"
)

const (
	Prefix = "tk+"
	Length = 64 // sha256的长度为32字节，转换为16进制字符串后为64个字符

	tk_regex_str = `^tk\+[a-fA-F0-9]{64}$`
)

var (
	token_regex *regexp.Regexp
)

// ValidateToken 检查令牌是否合法
func ValidateToken(str string) bool {

	if str == "" {
		return false
	}

	// 检查前缀是否正确
	if strings.HasPrefix(str, Prefix) == false {
		return false
	}

	// 检查长度是否正确
	str_len := len(str)
	if str_len != (Length + len(Prefix)) {
		return false
	}

	if !token_regex.MatchString(str) {
		return false
	}

	return true
}

func init() {

	token_regex = regexp.MustCompile(tk_regex_str)
	if token_regex == nil {
		panic("failed to compile token regex")
	}

}
