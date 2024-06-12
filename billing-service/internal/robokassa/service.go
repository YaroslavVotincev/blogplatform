package robokassa

import (
	"billing-service/internal/blogs"
	"billing-service/internal/posts"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/google/uuid"
	configService "github.com/llc-ldbit/go-cloud-config-client"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Service struct {
	repository   *Repository
	blogsService *blogs.Service
	postsService *posts.Service

	MerchantLogin string
	Password1     string
	Password2     string
	IsTest        bool
	TestPassword1 string
	TestPassword2 string
}

func NewService(repository *Repository, blogsService *blogs.Service, postsService *posts.Service,
	MerchantLogin, Password1, Password2, TestPassword1, TestPassword2 string, IsTest bool,
	cfgService *configService.ConfigServiceManager) *Service {

	service := Service{
		repository:    repository,
		blogsService:  blogsService,
		postsService:  postsService,
		MerchantLogin: MerchantLogin,
		Password1:     Password1,
		Password2:     Password2,
		IsTest:        IsTest,
		TestPassword1: TestPassword1,
		TestPassword2: TestPassword2,
	}

	cfgService.SetUpdateHandler(func(ss configService.ServiceSetting) {
		value, err := strconv.ParseBool(ss.Value)
		if err == nil {
			service.IsTest = value
		} else {
			service.IsTest = false
		}
	}, IsTestConfigKey)

	cfgService.SetUpdateHandler(func(ss configService.ServiceSetting) {
		service.MerchantLogin = ss.Value
	}, MerchantLoginConfigKey)

	cfgService.SetUpdateHandler(func(ss configService.ServiceSetting) {
		service.Password1 = ss.Value
	}, Password1ConfigKey)

	cfgService.SetUpdateHandler(func(ss configService.ServiceSetting) {
		service.Password2 = ss.Value
	}, Password2ConfigKey)

	cfgService.SetUpdateHandler(func(ss configService.ServiceSetting) {
		service.TestPassword1 = ss.Value
	}, TestPassword1ConfigKey)

	cfgService.SetUpdateHandler(func(ss configService.ServiceSetting) {
		service.TestPassword2 = ss.Value
	}, TestPassword2ConfigKey)

	return &service
}

func (s *Service) ConfirmSignatureValueValid(OutSum, InvId, SignatureValue string) bool {
	var sign string
	if s.IsTest {
		sign = fmt.Sprintf("%s:%s:%s", OutSum, InvId, s.TestPassword2)
		//fmt.Println(sign)
	} else {
		sign = fmt.Sprintf("%s:%s:%s", OutSum, InvId, s.Password2)
		//fmt.Println(sign)
	}
	sha := sha256.Sum256([]byte(sign))
	SignatureValue = strings.ToUpper(SignatureValue)
	CheckSum := strings.ToUpper(hex.EncodeToString(sha[:]))
	return SignatureValue == CheckSum
}

func (s *Service) CreateInvoice(ctx context.Context, itemId uuid.UUID, itemType, description string, sum float64, userId uuid.UUID) (string, error) {
	timeNow := time.Now().UTC()
	expirationDate := timeNow.Add(1 * time.Hour).UTC()
	pass := s.Password1
	if s.IsTest {
		pass = s.TestPassword1
	}
	invoice := Invoice{
		OutSum:    sum,
		ItemId:    itemId,
		ItemType:  itemType,
		UserId:    userId,
		ExpiresAt: expirationDate,
		Status:    InvoiceStatusNew,
		Created:   timeNow,
		Updated:   timeNow,
	}

	err := s.repository.CreateInvoice(ctx, &invoice)
	if err != nil {
		return "", fmt.Errorf("fail to create invoice cause %v", err)
	}

	invoiceResponse, err := ApiCreateInvoice(s.MerchantLogin, description, pass, sum, invoice.ID, expirationDate, s.IsTest)
	if err != nil {
		invoice.Status = InvoiceStatusFailed
		_ = s.repository.UpdateInvoice(ctx, &invoice)
		return "", fmt.Errorf("fail to create invoice through robokassa api cause %v", err)
	}

	fullUrl, err := url.JoinPath(PaymentLinkPrefixUrl, invoiceResponse.InvoiceId)
	if err != nil {
		invoice.Status = InvoiceStatusFailed
		_ = s.repository.UpdateInvoice(ctx, &invoice)
		return "", fmt.Errorf("fail to join url cause %v", err)
	}

	invoice.PaymentLink = fullUrl
	return fullUrl, s.repository.UpdateInvoice(ctx, &invoice)
}

func (s *Service) GrantItemByType(ctx context.Context, invoice *Invoice) error {

	if invoice.ItemType == InvoiceItemTypeSubscription {
		return s.blogsService.GrantSubscription(invoice.ItemId, invoice.UserId, invoice.OutSum, CurrencyRub)
	} else if invoice.ItemType == InvoiceItemTypePost {
		return s.postsService.GrantPostPaidAccess(invoice.ItemId, invoice.UserId, invoice.OutSum, CurrencyRub)
	} else if invoice.ItemType == InvoiceItemTypeDonation {
		return s.blogsService.ConfirmDonation(invoice.ItemId, invoice.UserId, invoice.OutSum, CurrencyRub)
	} else {
		return fmt.Errorf("unknown item type: %s", invoice.ItemType)
	}

}
