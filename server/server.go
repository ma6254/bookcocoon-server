package server

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/bwmarrin/snowflake"
	"github.com/ma6254/bookcocoon-server/config"
	"github.com/ma6254/bookcocoon-server/database"
	"github.com/ma6254/bookcocoon-server/server/tokens"
)

const (
	log_flags    = log.LstdFlags | log.Lshortfile // 日志格式
	token_header = "Authorization"                // HTTP请求头中用于传递token的字段名

	token_expire_time = 30 * time.Minute // token过期时间
)

type Server struct {
	srv      *http.Server
	Config   config.Config
	Mux      *http.ServeMux
	ExitChan chan struct{}
	log      *log.Logger
	DB       *database.Database
	// sessions       map[string]*Session // key: session_id
	sessions       sync.Map // key: session_id
	is_installed   bool
	start_time     time.Time
	snowflake_node *snowflake.Node
}

func NewServer(c *config.Config) *Server {

	snowflake_node, err := snowflake.NewNode(1)
	if err != nil {
		panic(err)
	}

	return &Server{
		Config:         *c,
		Mux:            http.NewServeMux(),
		ExitChan:       make(chan struct{}, 1),
		sessions:       sync.Map{},
		snowflake_node: snowflake_node,
	}
}

func (s *Server) Run() error {

	var (
		err error
	)

	s.log = log.New(os.Stdout, "[server]", log_flags)
	s.start_time = time.Now()

	is_installed := s.IsInstalled()

	// 初始化数据库
	s.DB, err = database.NewDatabase(&s.Config.DataBase)
	if err != nil {
		return err
	}
	s.log.Printf("database: %s\r\n", s.Config.DataBase.Driver)

	if is_installed == false {
		s.log.Printf("Server is not installed, now installing...\r\n")
		err = s.Install()
		if err != nil {
			s.log.Printf("install failed\r\n")
			return err
		}
	}

	s.log.Printf("http server: http://%s\r\n", s.Config.Server.HttpAddr)
	s.log.Printf("swagger UI: http://%s/swagger/index.html\r\n", s.Config.Server.HttpAddr)
	s.log.Printf("swagger spec: http://%s/docs/swagger.yaml\r\n", s.Config.Server.HttpAddr)

	err = s.setRoute()
	if err != nil {
		return err
	}

	err = s.createAdminUser()
	if err != nil {
		return err
	}

	s.srv = &http.Server{
		Addr:    s.Config.Server.HttpAddr,
		Handler: s.Mux,
	}

	// // 获取端口号
	// _, port_str, err := net.SplitHostPort(s.Config.Server.HttpAddr)
	// if err != nil {
	// 	s.log.Printf("Failed to get port: %v", err)
	// 	return err
	// }

	// port, _ := strconv.ParseInt(port_str, 10, 32)

	// ok, pid, thread_name := utils.GetPidByTCPPort(int(port))
	// if ok {
	// 	return fmt.Errorf("port %d is already in use by process %d (%s)", port, pid, thread_name)
	// }

	go func() {

		var (
			err error
		)

		// 确保在服务器退出时发送退出信号
		defer func() {
			s.log.Printf("Server exited")
			s.ExitChan <- struct{}{}
		}()

		// 启动HTTP服务器
		s.log.Printf("Starting server on %s", s.Config.Server.HttpAddr)
		err = s.srv.ListenAndServe()
		if err == nil {
			return
		}

		// 处理服务器关闭的情况
		if err == http.ErrServerClosed {
			s.log.Printf("server normally shutdown")
			return
		}

		// 处理其他错误
		s.log.Printf("Failed to start server: %v", err)
	}()

	return nil
}

func (s *Server) Stop() error {

	// 关闭所有会话
	s.sessions.Range(func(key, value any) bool {
		session := value.(*Session)
		session.Delete()
		s.sessions.Delete(key)
		return true
	})

	// 关闭数据库连接
	if s.DB != nil {
		s.DB.Close()
	}

	// 关闭HTTP服务器
	if s.srv != nil {
		err := s.srv.Close()
		if err != nil {
			s.log.Printf("Error shutting down server: %v", err)
			return err
		}
	}

	return nil
}

// Salt 将用户名和密码进行加盐处理，生成一个唯一的字符串
func (s *Server) Salt(user_id uint64, password string) string {

	src_str := s.Config.DataBase.Salt + strconv.FormatUint(user_id, 10) + password

	h := sha256.New()
	h.Write([]byte(src_str))
	src_str = fmt.Sprintf("%x", h.Sum(nil))

	return src_str
}

// 生成一个随机的会话ID
func (s *Server) GenerateSessionID() string {

	id, err := uuid.NewV7()
	if err != nil {
		panic(err)
	}

	return fmt.Sprintf("%s", id.String())
}

// 生成一个随机的token
func (s *Server) GenerateToken() string {

	tk := "tk+"

	id, err := uuid.NewV7()
	if err != nil {
		panic(err)
	}

	src_str := ""
	src_str += fmt.Sprintf("%s", id.String())
	src_str += fmt.Sprintf("%d", time.Now().UnixNano())

	h := sha256.New()
	h.Write([]byte(src_str))

	return tk + fmt.Sprintf("%x", h.Sum(nil))
}

// NewSession 建立新的会话
func (s *Server) NewSession(user_name string, user_id uint64, token string) (session *Session, exists bool, err error) {

	session = s.FindSessionByToken(token)
	if session != nil {
		return session, true, nil
	}

	session_id := s.GenerateSessionID()
	session_log := log.New(os.Stdout, fmt.Sprintf("[session][%s]", user_name), log_flags)
	session = NewSession(
		user_name,
		user_id,
		token,
		session_id,
		session_log)
	s.sessions.Store(session.SessionID, session)

	go func() {

		exit := false

		for {
			select {
			case <-session.exit_signal:
				s.log.Printf("[session][%d] exit\r\n", session.UserID)
				exit = true
			case <-session.heartbeat_timer.C:
				s.log.Printf("[session][%d] heartbeat timeout\r\n", session.UserID)
				exit = true
			}

			if exit {
				break
			}

		}

		session.Delete()
		s.sessions.Delete(session.SessionID)
		s.log.Printf("[session][%d] session deleted\r\n", session.UserID)

	}()

	return session, false, nil
}

// FindSession 查找会话
func (s *Server) FindSession(session_id string) (session *Session) {

	if session_id == "" {
		return nil
	}

	value, exists := s.sessions.Load(session_id)
	if exists {
		session = value.(*Session)
		return session
	}

	return nil
}

func (s *Server) FindSessionByToken(token string) (session *Session) {

	if token == "" {
		return nil
	}

	session = nil

	s.sessions.Range(func(key, value any) bool {
		curr_session := value.(*Session)
		if curr_session.Token == token {
			session = curr_session
			return false
		}
		return true
	})

	return session
}

func (s *Server) createAdminUser() error {

	var (
		err         error
		ok          bool
		admin_user  = "admin"
		admin_token = "tk+3f3bc6d699c048fa9c8cf6d46ffc80b39d75db71b9183846bd30d72fc618ca71"
	)

	s.log.Printf("create admin user: %s\r\n", admin_user)

	token_valid := tokens.ValidateToken(admin_token)
	if token_valid == false {
		return fmt.Errorf("admin token is not valid")
	}

	ok, err = s.DB.CheckToken(admin_token, admin_user_id)
	if err != nil {
		return fmt.Errorf("create admin token failed: %s", err.Error())
	}

	if ok == false {

		s.log.Printf("admin token does not exist, create\r\n")

		_, err = s.DB.CreateToken(admin_token, admin_user_id)
		if err != nil {
			return fmt.Errorf("create admin token failed: %s", err.Error())
		}
	}

	_, _, err = s.NewSession(admin_user, admin_user_id, admin_token)
	if err != nil {
		return fmt.Errorf("create admin session failed: %s", err.Error())
	}

	s.log.Printf("create admin session success\r\n")
	return nil
}

// WriteJsonSuccessResponse 将数据写入HTTP响应，格式为JSON，并返回状态码 200_OK
func (s *Server) WriteJsonSuccessResponse(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")

	body_buf := bytes.NewBuffer(nil)
	encoder := json.NewEncoder(body_buf)
	encoder.SetIndent("", " ")

	err := encoder.Encode(data)
	if err != nil {
		http.Error(w, "json encode error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(body_buf.Bytes())
}
