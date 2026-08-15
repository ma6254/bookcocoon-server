package validator

import (
	"fmt"
	"regexp"
)

const (
	user_name_regex_str = "^[a-zA-Z][a-zA-Z0-9_-]{2,15}$" // 用户名正则表达式
	user_name_min_len   = 3                               // 用户名最小长度
	user_name_max_len   = 256                             // 用户名最大长度

	user_email_regex_str = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$` // 邮箱正则表达式
	user_email_min_len   = 3                                                  // 邮箱最小长度
	user_email_max_len   = 256                                                // 邮箱最大长度

	user_passwd_regex_str = `^[a-zA-Z0-9._%+-]{8,}$` // 密码正则表达式
	user_passwd_min_len   = 8                        // 密码最小长度
	user_passwd_max_len   = 256                      // 密码最大长度
)

var (
	username_regex    *regexp.Regexp
	user_email_regex  *regexp.Regexp
	user_passwd_regex *regexp.Regexp
)

// ValidateUserName 验证用户名是否合法
func ValidateUserName(input string) error {

	// 检查用户名长度是否在允许范围内
	input_len := len(input)
	if input_len < user_name_min_len || input_len > user_name_max_len {
		return fmt.Errorf("username length must be between %d and %d", user_name_min_len, user_name_max_len)
	}

	// 检查用户名是否符合正则表达式
	if !username_regex.MatchString(input) {
		return fmt.Errorf("username does not match the required pattern")
	}

	return nil
}

// ValidateEmail 验证邮箱是否合法
func ValidateUserEmail(input string) error {

	// 检查邮箱长度是否在允许范围内
	input_len := len(input)
	if input_len < user_email_min_len || input_len > user_email_max_len {
		return fmt.Errorf("email length must be between %d and %d", user_email_min_len, user_email_max_len)
	}

	// 检查邮箱是否符合正则表达式
	if !user_email_regex.MatchString(input) {
		return fmt.Errorf("email does not match the required pattern")
	}

	return nil
}

// ValidateUserPassword 验证密码是否合法
func ValidateUserPassword(input string) error {

	// 检查密码长度是否在允许范围内
	input_len := len(input)
	if input_len < user_passwd_min_len || input_len > user_passwd_max_len {
		return fmt.Errorf("password length must be between %d and %d", user_passwd_min_len, user_passwd_max_len)
	}

	// 检查密码是否符合正则表达式
	matched := user_passwd_regex.MatchString(input)
	if !matched {
		return fmt.Errorf("password does not match the required pattern")
	}

	return nil
}

func init() {
	username_regex = regexp.MustCompile(user_name_regex_str)
	if username_regex == nil {
		panic("failed to compile username regex")
	}

	user_email_regex = regexp.MustCompile(user_email_regex_str)
	if user_email_regex == nil {
		panic("failed to compile user email regex")
	}

	user_passwd_regex = regexp.MustCompile(user_passwd_regex_str)
	if user_passwd_regex == nil {
		panic("failed to compile user password regex")
	}
}
