package server

import (
	"log"
	"time"
)

const (
	enable_token_expire = true
)

type Session struct {
	UserID          uint64
	UserName        string
	Token           string
	SessionID       string
	log             *log.Logger
	heartbeat_timer *time.Timer
	exit_signal     chan struct{}
}

// NewSession 创建一个新的会话
func NewSession(user_name string, user_id uint64, token string, session_id string, logger *log.Logger) *Session {

	var heartbeat_timer *time.Timer

	if enable_token_expire {
		heartbeat_timer = time.NewTimer(token_expire_time)
	}

	s := &Session{
		UserID:          user_id,
		UserName:        user_name,
		SessionID:       session_id,
		Token:           token,
		log:             logger,
		heartbeat_timer: heartbeat_timer,
		exit_signal:     make(chan struct{}, 1),
	}

	return s
}

// Delete 删除会话
func (s *Session) Delete() error {

	select {
	case s.exit_signal <- struct{}{}:
	default:
	}

	if s.heartbeat_timer != nil {
		s.heartbeat_timer.Stop()
	}

	return nil
}
