package users

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"log"
	"net/url"
	"time"
	"users-service/internal/files"
	"users-service/internal/wallet"
	"users-service/pkg/cryptservice"
)

const (
	DefaultAdminLogin    = "admin"
	DefaultAdminEmail    = "admin@admin.com"
	DefaultAdminPassword = "admin"
)

type Service struct {
	repository    *Repository
	walletService *wallet.Service
	filesService  *files.Service
}

func NewService(ctx context.Context, repository *Repository, walletService *wallet.Service, filesService *files.Service) *Service {
	service := Service{
		repository:    repository,
		walletService: walletService,
		filesService:  filesService,
	}
	service.init(ctx)
	return &service
}

func (s *Service) init(ctx context.Context) {
	count, err := s.repository.Count(ctx)
	if err != nil {
		log.Fatalf("failed to get all users count cause %v", err)
	}
	if count == 0 {
		log.Println("no users found. creating default admin")
		_, err := s.Create(ctx,
			DefaultAdminLogin,
			DefaultAdminEmail,
			DefaultAdminPassword,
			"admin",
			true,
			nil,
		)

		if err != nil {
			log.Fatalf("failed to create default admin user cause %v", err)
		}
	}
}

func (s *Service) All(ctx context.Context) ([]User, error) {
	return s.repository.All(ctx)
}

func (s *Service) ByID(ctx context.Context, id uuid.UUID) (*User, error) {
	return s.repository.ByID(ctx, id)
}

func (s *Service) ByLogin(ctx context.Context, login string) (*User, error) {
	return s.repository.ByLogin(ctx, login)
}

func (s *Service) ByEmail(ctx context.Context, email string) (*User, error) {
	return s.repository.ByEmail(ctx, email)
}

func (s *Service) Create(ctx context.Context, login, email, password, role string, enabled bool, eraseAt *time.Time) (*User, error) {
	timeNow := time.Now().UTC()
	id := uuid.New()
	passwordHashed, err := cryptservice.CryptValue(password)
	if err != nil {
		return nil, err
	}
	user := User{
		ID:               id,
		Login:            login,
		Email:            email,
		HashedPassword:   passwordHashed,
		Role:             role,
		Deleted:          false,
		Enabled:          enabled,
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
	return &user, s.repository.Create(ctx, &user, &profile)
}

func (s *Service) Update(ctx context.Context, user *User, login, email, role string,
	deleted, enabled bool,
	bannedUntil *time.Time, bannedReason *string) error {

	timeNow := time.Now().UTC()
	user.Login = login
	user.Email = email
	user.Role = role
	user.Deleted = deleted
	user.Enabled = enabled
	user.Updated = timeNow
	user.BannedUntil = bannedUntil
	user.BannedReason = bannedReason
	return s.repository.Update(ctx, user)
}

func (s *Service) EraseById(ctx context.Context, id uuid.UUID) error {
	return s.repository.EraseById(ctx, id)
}

func (s *Service) WalletByUserId(ctx context.Context, id uuid.UUID) (*Wallet, error) {
	return s.repository.WalletByUserId(ctx, id)
}

func (s *Service) GetWalletBalance(walletAddress string) (float64, error) {
	return s.walletService.GetBalance(walletAddress)
}

func (s *Service) CreateWalletToUser(ctx context.Context, userId uuid.UUID) (*Wallet, error) {
	walletResponse, err := s.walletService.CreateWallet()
	if err != nil {
		return nil, fmt.Errorf("wallet service error: %v", err)
	}
	walletObj := Wallet{
		ID:        userId,
		PublicKey: walletResponse.PublicKey,
		SecretKey: walletResponse.SecretKey,
		Mnemonic:  walletResponse.Mnemonic,
		Address:   walletResponse.Address,
		Created:   time.Now().UTC(),
	}

	return &walletObj, s.repository.CreateWallet(ctx, &walletObj)
}

func (s *Service) ProfileById(ctx context.Context, userId uuid.UUID) (*Profile, error) {
	return s.repository.ProfileById(ctx, userId)
}

func (s *Service) UpdateProfileFromFioRequest(ctx context.Context, profile *Profile, fioRequest *MyProfileFioRequest) error {
	profile.FirstName = fioRequest.FirstName
	profile.LastName = fioRequest.LastName
	profile.MiddleName = fioRequest.MiddleName
	return s.repository.UpdateProfile(ctx, profile)
}

func (s *Service) ByIdHashedPassword(ctx context.Context, id uuid.UUID) (*string, error) {
	return s.repository.ByIdHashedPassword(ctx, id)
}

func (s *Service) UpdatePassword(ctx context.Context, userId uuid.UUID, newPassword string) error {
	passwordHashed, err := cryptservice.CryptValue(newPassword)
	if err != nil {
		return err
	}
	return s.repository.UpdatePassword(ctx, userId, passwordHashed)
}

func (s *Service) RedirectUrlToAvatar(avatar string) (string, error) {
	path, err := url.JoinPath(s.filesService.GetFileEndpointUrl(), avatar)
	if err != nil {
		return "", err
	}
	return path, nil
}

func (s *Service) SetProfileAvatar(ctx context.Context, profile *Profile, bytes []byte) error {
	if profile.Avatar == nil {
		id := uuid.New().String()
		profile.Avatar = &id
		err := s.repository.UpdateProfile(ctx, profile)
		if err != nil {
			return err
		}
	}
	go s.filesService.SendFile(*profile.Avatar, bytes)
	return nil
}
