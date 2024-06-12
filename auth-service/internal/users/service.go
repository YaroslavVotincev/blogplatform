package users

import (
	"auth-service/internal/notifications"
	"auth-service/pkg/filelogger"
	"auth-service/pkg/queuelogger"
	"context"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	configService "github.com/llc-ldbit/go-cloud-config-client"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	JwtSecretConfigKey             = "JWT_SECRET"
	JwtDefaultLifetimeConfigKey    = "JWT_DEFAULT_LIFETIME"
	JwtRememberMeLifetimeConfigKey = "JWT_REMEMBER_ME_LIFETIME"
)

type Service struct {
	repository       *Repository
	notifService     *notifications.Service
	jwtSecret        []byte
	defaultLifetime  int
	rememberLifetime int
	mu               *sync.RWMutex
}

func NewService(repository *Repository, notifService *notifications.Service,
	jwtSecret string, defaultLifetime int, rememberLifetime int,
	cfgService *configService.ConfigServiceManager,
	fileLogger *filelogger.FileLogger,
	queueLogger *queuelogger.RemoteLogger) *Service {

	service := Service{
		repository:       repository,
		notifService:     notifService,
		jwtSecret:        []byte(jwtSecret),
		defaultLifetime:  defaultLifetime,
		rememberLifetime: rememberLifetime,
		mu:               &sync.RWMutex{},
	}

	cfgService.SetUpdateHandler(func(ss configService.ServiceSetting) {
		service.mu.Lock()
		service.jwtSecret = []byte(ss.Value)
		service.mu.Unlock()
	}, JwtSecretConfigKey)

	cfgService.SetUpdateHandler(func(ss configService.ServiceSetting) {
		value, err := strconv.Atoi(ss.Value)
		if err != nil {
			data := map[string]any{"error": err.Error(), "key": ss.Key, "value": ss.Value}
			fileLogger.Error("failed to parse value for jwt default lifetime from config", data)
			data["message"] = "failed to parse value for jwt default lifetime from config"
			err1 := queueLogger.Error(nil, data)
			if err1 != nil {
				fileLogger.Error("failed to send log about jwt default lifetime parsing error to queue",
					map[string]any{"error": err1.Error()})
			}
		}
		service.mu.Lock()
		service.defaultLifetime = value
		service.mu.Unlock()
	}, JwtDefaultLifetimeConfigKey)

	cfgService.SetUpdateHandler(func(ss configService.ServiceSetting) {
		value, err := strconv.Atoi(ss.Value)
		if err != nil {
			data := map[string]any{"error": err.Error(), "key": ss.Key, "value": ss.Value}
			fileLogger.Error("failed to parse value for jwt remember lifetime from config", data)
			data["message"] = "failed to parse value for jwt remember lifetime from config"
			err1 := queueLogger.Error(nil, data)
			if err1 != nil {
				fileLogger.Error("failed to send log about jwt remember lifetime parsing error to queue",
					map[string]any{"error": err1.Error()})
			}
		}
		service.mu.Lock()
		service.rememberLifetime = value
		service.mu.Unlock()
	}, JwtRememberMeLifetimeConfigKey)

	return &service
}

func (s *Service) GetJwtSecret() []byte {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.jwtSecret
}

func (s *Service) GetLifetime(remember bool) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if remember {
		return s.rememberLifetime
	}
	return s.defaultLifetime
}

func (s *Service) ByID(ctx context.Context, id uuid.UUID) (*User, error) {
	return s.repository.ByID(ctx, id)
}

func (s *Service) ByValue(ctx context.Context, value string) (*User, error) {
	return s.repository.ByValue(ctx, value)
}

func (s *Service) GenerateToken(userId uuid.UUID, remember bool) (string, error) {
	exp := time.Now().Add(time.Hour * time.Duration(s.GetLifetime(remember))).Unix()
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userId.String(),
		"exp":     exp,
	})
	return t.SignedString(s.GetJwtSecret())
}

func (s *Service) KeyFunction(token *jwt.Token) (interface{}, error) {
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
	}
	return s.GetJwtSecret(), nil
}

func (s *Service) ParseToken(authorizationHeader string) (*uuid.UUID, error) {
	if authorizationHeader == "" {
		return nil, fmt.Errorf("authorization header is empty")
	}
	headerParts := strings.Split(authorizationHeader, " ")
	if len(headerParts) != 2 {
		return nil, fmt.Errorf("invalid authorization header")
	}
	if headerParts[0] != "Bearer" {
		return nil, fmt.Errorf("invalid authorization header")
	}

	token, err := jwt.Parse(headerParts[1], s.KeyFunction)
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userId, err := uuid.Parse(claims["user_id"].(string))
		if err != nil {
			return nil, err
		}
		return &userId, nil
	} else {
		return nil, err
	}
}
