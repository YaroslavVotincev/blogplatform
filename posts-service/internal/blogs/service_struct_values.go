package blogs

import (
	configService "posts-service/pkg/config-client"
	"strconv"
)

func (s *Service) SetConfigUpdateHandlers(cfgService *configService.ConfigServiceManager) {
	cfgService.SetUpdateHandler(func(ss configService.ServiceSetting) {
		value, err := strconv.Atoi(ss.Value)
		if err == nil {
			s.mu.Lock()
			s.mainPageLikesRequirement = value
			s.mu.Unlock()
		}
	}, "MAIN_PAGE_LIKES_REQUIREMENT")
	cfgService.SetUpdateHandler(func(ss configService.ServiceSetting) {
		value, err := strconv.Atoi(ss.Value)
		if err == nil {
			s.mu.Lock()
			s.mainPageCommentsRequirement = value
			s.mu.Unlock()
		}
	}, "MAIN_PAGE_COMMENTS_REQUIREMENT")
	cfgService.SetUpdateHandler(func(ss configService.ServiceSetting) {
		value, err := strconv.Atoi(ss.Value)
		if err == nil {
			s.mu.Lock()
			s.mainPageViewsRequirement = value
			s.mu.Unlock()
		}
	}, "MAIN_PAGE_VIEWS_REQUIREMENT")
	cfgService.SetUpdateHandler(func(ss configService.ServiceSetting) {
		value, err := strconv.Atoi(ss.Value)
		if err == nil {
			s.mu.Lock()
			s.mainPageDislikesRequirement = value
			s.mu.Unlock()
		}
	}, "MAIN_PAGE_DISLIKES_REQUIREMENT")

	cfgService.SetUpdateHandler(func(ss configService.ServiceSetting) {
		value, err := strconv.ParseFloat(ss.Value, 64)
		if err == nil {
			s.mu.Lock()
			s.donationsRobokassaMinValue = value
			s.mu.Unlock()
		}
	}, "DONATIONS_ROBOKASSA_MIN_VALUE")

	cfgService.SetUpdateHandler(func(ss configService.ServiceSetting) {
		value, err := strconv.ParseFloat(ss.Value, 64)
		if err == nil {
			s.mu.Lock()
			s.donationsToncoinMinValue = value
			s.mu.Unlock()
		}
	}, "DONATIONS_TONCOIN_MIN_VALUE")
}

func (s *Service) MainPageLikesRequirement() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.mainPageLikesRequirement
}

func (s *Service) MainPageCommentsRequirement() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.mainPageCommentsRequirement
}

func (s *Service) MainPageViewsRequirement() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.mainPageViewsRequirement
}

func (s *Service) MainPageDislikesRequirement() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.mainPageDislikesRequirement
}

func (s *Service) DonationsRobokassaMinValue() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.donationsRobokassaMinValue
}

func (s *Service) DonationsToncoinMinValue() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.donationsToncoinMinValue
}
