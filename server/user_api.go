package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/ma6254/bookcocoon-server/database"
	"github.com/ma6254/bookcocoon-server/validator"
)

var (
	HttpErrorUnauthorized = func(w http.ResponseWriter) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	}
)

// LoginForm 用户登录表单
type LoginForm struct {
	Account  string `json:"account"`  // 用户ID、用户名、邮箱
	Password string `json:"password"` // 用户密码
}

// LoginResponse 用户登录响应
type LoginResponse struct {
	Token    string    `json:"token"`
	UserInfo *UserInfo `json:"user_info"`
}

type UserInfo struct {
	UserID    string `json:"user_id"`    // 用户ID
	UserName  string `json:"user_name"`  // 用户名
	CreatedAt string `json:"created_at"` // 创建时间
	LoginAt   string `json:"login_at"`   // 最近登录时间
	NickName  string `json:"nick_name"`  // 昵称
	Email     string `json:"email"`      // 邮箱
}

func NewUserInfoByDB(db_auth *database.Auth, db_user *database.User) *UserInfo {

	return &UserInfo{
		UserID:    fmt.Sprintf("%d", db_user.ID),
		UserName:  db_user.UserName,
		CreatedAt: db_auth.CreatedAt,
		LoginAt:   db_auth.LoginAt,
		NickName:  db_user.NickName,
		Email:     db_user.Email,
	}
}

// login_handler 登录回调处理
// @Summary      用户登录
// @Description  用户登录
// @Tags         账户
// @Accept       json
// @Produce      json
// @Param        login_form  body      LoginForm  true  "登录表单"
// @Success      200  {object}  LoginResponse
// @Router       /user/login [post]
func (s *Server) http_api_login_handler() func(w http.ResponseWriter, r *http.Request) {

	log_prefix := fmt.Sprintf("%s[login]", s.log.Prefix())
	curr_log := log.New(os.Stdout, log_prefix, log_flags)

	return func(w http.ResponseWriter, r *http.Request) {

		var (
			err        error
			login_form = LoginForm{}
		)

		err = json.NewDecoder(r.Body).Decode(&login_form)
		if err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		if (login_form.Account == "") || (login_form.Password == "") {
			http.Error(w, "account and password must not be empty", http.StatusBadRequest)
			return
		}

		err = validator.ValidateUserPassword(login_form.Password)
		if err != nil {
			curr_log.Printf("validate password failed: %s\r\n", err.Error())
			http.Error(w, "validate password failed", http.StatusBadRequest)
			return
		}

		var user *database.User
		user, err = s.DB.CheckUserAccount(login_form.Account)
		if err != nil {
			curr_log.Printf("check user account failed: %s\r\n", err.Error())
			http.Error(w, "check user account failed", http.StatusInternalServerError)
			return
		}

		if user == nil {
			HttpErrorUnauthorized(w)
			return
		}

		hashed_password := s.Salt(user.ID, login_form.Password)

		// curr_log.Printf("account: %#v, password: %#v, hashed_password: %#v", login_form.Account, login_form.Password, hashed_password)

		// 查询数据库，检查用户名和密码是否正确
		auth, err := s.DB.UserAuth(user.ID, hashed_password)
		if err != nil {
			HttpErrorInternal(w, err)
			return
		}
		if auth == nil {
			HttpErrorUnauthorized(w)
			return
		}

		curr_log.Printf("account: %#v, id: %d, user_name: %#v, email: %#v", login_form.Account, user.ID, user.UserName, user.Email)

		// 查找是否已经存在会话，如果存在则删除旧的会话
		session_exists := false
		token := ""

		s.sessions.Range(func(key, value any) bool {
			session := value.(*Session)
			if session.UserName == user.UserName {

				session_exists = true
				token = session.Token
				return false
			}
			return true
		})

		// 如果已经存在会话，则直接返回登录成功的响应
		if session_exists {

			curr_log.Printf("session_exists name:%#v", user.UserName)

			// 返回登录成功的响应
			s.WriteJsonSuccessResponse(w, LoginResponse{
				Token:    token,
				UserInfo: NewUserInfoByDB(auth, user),
			})
			return
		}

		// 如果不存在会话，则查询数据库，检查是否已经存在token
		tokens, err := s.DB.GetUserAllAliveToken(user.ID)
		if err != nil {
			HttpErrorInternal(w, err)
			return
		}

		// 如果已经存在token，则直接返回登录成功的响应
		if len(tokens) > 0 {
			token = tokens[0].Token

			curr_log.Printf("token_exist new_session name:%#v\r\n", user.UserName)

			// 创建新的会话
			_, _, err = s.NewSession(user.UserName, auth.ID, token)
			if err != nil {
				HttpErrorInternal(w, err)
				return
			}

			// 返回登录成功的响应
			s.WriteJsonSuccessResponse(w, LoginResponse{
				Token:    token,
				UserInfo: NewUserInfoByDB(auth, user),
			})
			return
		}

		// 如果不存在会话和token，则生成新的token，并创建新的会话
		token = s.GenerateToken()

		curr_log.Printf("new_token new_session name:%#v\r\n", user.UserName)

		_, err = s.DB.CreateToken(token, auth.ID)
		if err != nil {
			HttpErrorInternal(w, err)
			return
		}

		// 创建新的会话
		_, _, err = s.NewSession(user.UserName, auth.ID, token)
		if err != nil {
			HttpErrorInternal(w, err)
			return
		}

		// 返回登录成功的响应
		s.WriteJsonSuccessResponse(w, LoginResponse{
			Token:    token,
			UserInfo: NewUserInfoByDB(auth, user),
		})
	}
}
