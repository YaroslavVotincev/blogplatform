package users

import (
	"context"
	"github.com/google/uuid"
	"github.com/llc-ldbit/go-cloud-config-client"
	"registration-service/internal/email"
	"registration-service/internal/notifications"
	"registration-service/pkg/cryptservice"
	"registration-service/pkg/filelogger"
	requestuser "registration-service/pkg/hidepost-requestuser"
	"registration-service/pkg/queuelogger"
	"strconv"
	"time"
)

const (
	UserAutoEnableConfigKey      = "USER_AUTO_ENABLE"
	UserCleanupIntervalConfigKey = "USER_CLEANUP_INTERVAL"
	SignupLifetimeConfigKey      = "SIGNUP_LIFETIME"
)

type Service struct {
	repository          *Repository
	emailService        *email.Service
	notifService        *notifications.Service
	autoEnable          bool
	signupLifetime      int
	userCleanupInterval int
}

func NewService(ctx context.Context, repository *Repository,
	emailService *email.Service, notifService *notifications.Service,
	autoEnable bool, signupLifetime int, userCleanupInterval int,
	cfgService *configService.ConfigServiceManager,
	fileLogger *filelogger.FileLogger,
	queueLogger *queuelogger.RemoteLogger) *Service {

	service := Service{
		repository:          repository,
		emailService:        emailService,
		notifService:        notifService,
		autoEnable:          autoEnable,
		signupLifetime:      signupLifetime,
		userCleanupInterval: userCleanupInterval,
	}

	cfgService.SetUpdateHandler(func(ss configService.ServiceSetting) {
		value, err := strconv.ParseBool(ss.Value)
		if err == nil {
			service.autoEnable = value
		}
	}, UserAutoEnableConfigKey)

	cfgService.SetUpdateHandler(func(ss configService.ServiceSetting) {
		value, err := strconv.Atoi(ss.Value)
		if err != nil {
			data := map[string]any{"error": err.Error(), "key": ss.Key, "value": ss.Value}
			fileLogger.Error("failed to parse value for signup lifetime from config", data)
			data["message"] = "failed to parse value for signup lifetime from config"
			err1 := queueLogger.Error(nil, data)
			if err1 != nil {
				fileLogger.Error("failed to send log about signup lifetime parsing error to queue",
					map[string]any{"error": err1.Error()})
			}
		}
		service.signupLifetime = value
	}, SignupLifetimeConfigKey)

	cfgService.SetUpdateHandler(func(ss configService.ServiceSetting) {
		interval, err := strconv.Atoi(ss.Value)
		if err != nil {
			data := map[string]any{"error": err.Error(), "key": ss.Key, "value": ss.Value}
			fileLogger.Error("failed to parse value for user cleanup interval from config", data)
			data["message"] = "failed to parse value for user cleanup interval from config"
			err1 := queueLogger.Error(nil, data)
			if err1 != nil {
				fileLogger.Error("failed to send log about user cleanup interval parsing error to queue",
					map[string]any{"error": err1.Error()})
			}
		}
		service.userCleanupInterval = interval
	}, UserCleanupIntervalConfigKey)

	go func() {
		for {
			err := service.repository.EraseAllDue(ctx)
			if err != nil {
				data := map[string]any{"error": err.Error()}
				fileLogger.Error("failed to erase all due users", data)
				data["message"] = "failed to erase all due users"
				err1 := queueLogger.Error(nil, data)
				if err1 != nil {
					fileLogger.Error("failed to send log about user cleanup to queue",
						map[string]any{"error": err1.Error()})
				}
			}
			time.Sleep(time.Duration(service.userCleanupInterval) * time.Hour)
		}
	}()

	return &service
}

func (s *Service) ExistsById(ctx context.Context, id uuid.UUID) (bool, error) {
	return s.repository.ExistsById(ctx, id)
}

func (s *Service) ExistsByLogin(ctx context.Context, login string) (bool, error) {
	return s.repository.ExistsByLogin(ctx, login)
}

func (s *Service) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	return s.repository.ExistsByEmail(ctx, email)
}

func (s *Service) HandleSignup(ctx context.Context, login, email, password string) error {
	timeNow := time.Now().UTC()
	id := uuid.New()
	passwordHashed, err := cryptservice.CryptValue(password)
	if err != nil {
		return err
	}
	var eraseAt *time.Time = nil
	if s.autoEnable == false {
		temp := timeNow.Add(time.Duration(s.signupLifetime) * time.Hour)
		eraseAt = &temp
	}
	user := User{
		ID:               id,
		Login:            login,
		Email:            email,
		HashedPassword:   passwordHashed,
		Role:             requestuser.UserRoleUser,
		Deleted:          false,
		Enabled:          s.autoEnable,
		EmailConfirmedAt: nil,
		EraseAt:          eraseAt,
		Created:          timeNow,
		Updated:          timeNow,
		BannedUntil:      nil,
		BannedReason:     nil,
	}
	profile := Profile{
		ID:         id,
		FirstName:  "",
		LastName:   "",
		MiddleName: "",
	}

	err = s.repository.CreateUser(ctx, &user, &profile)
	if err != nil {
		return err
	}

	if s.autoEnable == false {
		code := uuid.New().String()
		err = s.repository.CreateConfirmationCode(ctx, code, id, timeNow.Add(time.Duration(s.signupLifetime)*time.Hour))
		if err != nil {
			return err
		}
		go s.emailService.SendSignupConfirmEmail(login, email, code)
	}
	return nil
}

func (s *Service) UserIdByConfirmCode(ctx context.Context, code string) (*uuid.UUID, error) {
	return s.repository.UserIdByConfirmCode(ctx, code)
}

func (s *Service) Enable(ctx context.Context, userId uuid.UUID, confirmCode string) error {
	return s.repository.Enable(ctx, userId, confirmCode)
}

func (s *Service) EraseConfirmCode(ctx context.Context, code string) error {
	return s.repository.EraseConfirmCode(ctx, code)
}
