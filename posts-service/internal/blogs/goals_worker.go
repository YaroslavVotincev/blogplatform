package blogs

import (
	"context"
	"github.com/google/uuid"
	"log"
	"math"
	"time"
)

func (s *Service) StartGoalsWorker(ctx context.Context, ticker *time.Ticker) {
	for range ticker.C {
		goals, _ := s.repository.AllGoals(ctx)
		for _, goal := range goals {
			err := s.HandleGoal(ctx, &goal)
			if err != nil {
				log.Println(err)
			}
		}
	}
}

func (s *Service) HandleGoal(ctx context.Context, goal *Goal) error {

	switch goal.Type {
	case "1":
		follows, err := s.repository.CountBlogUserFollows(ctx, goal.BlogId)
		if err != nil {
			return err
		}
		goal.Current = follows
		return s.repository.UpdateGoal(ctx, goal)
	case "2":
		subs, err := s.repository.CountBlogPaidUserSubscriptions(ctx, goal.BlogId)
		if err != nil {
			return err
		}
		goal.Current = subs
		return s.repository.UpdateGoal(ctx, goal)
	case "3":
		incomes, err := s.repository.BlogIncomesByBlogId(ctx, goal.BlogId)
		if err != nil {
			return err
		}
		var current float64 = 0
		for _, income := range incomes {
			if income.Currency == CurrencyRub {
				current += income.Value
			}
		}
		goal.Current = int(math.Round(current))
		return s.repository.UpdateGoal(ctx, goal)
	case "4":
		blogSubscriptions, err := s.repository.SubscriptionsByBlogId(ctx, goal.BlogId)
		if err != nil {
			return err
		}
		userSubsByBlog, err := s.repository.UserSubscriptionsByBlogId(ctx, goal.BlogId)
		if err != nil {
			return err
		}
		blogSubscriptionsMap := make(map[uuid.UUID]Subscription, len(blogSubscriptions))
		for _, sub := range blogSubscriptions {
			blogSubscriptionsMap[sub.ID] = sub
		}
		var current float64 = 0
		for _, userSub := range userSubsByBlog {
			if userSub.IsActive {
				if sub, ok := blogSubscriptionsMap[userSub.SubscriptionId]; ok {
					if sub.IsFree == false {
						current += sub.PriceRub
					}
				}
			}
		}

		goal.Current = int(math.Round(current))

		return s.repository.UpdateGoal(ctx, goal)

	default:
		return nil
	}
}

//func (s *Service) GoalType3AddSubscriptionValue(ctx context.Context, subscription *Subscription) error {
//	goals, err := s.repository.GoalsByBlogId(ctx, subscription.BlogId)
//	if err != nil {
//		return err
//	}
//	for _, goal := range goals {
//		if goal.Type == "3" {
//			goal.Current += int(math.Round(subscription.PriceRub))
//			err = s.repository.UpdateGoal(ctx, &goal)
//		}
//	}
//	return err
//}
//
//func (s *Service) GoalType3AddPostValue(ctx context.Context, post *Post) error {
//	goals, err := s.repository.GoalsByBlogId(ctx, post.BlogId)
//	if err != nil {
//		return err
//	}
//	for _, goal := range goals {
//		if goal.Type == "3" {
//			var value float64 = 0
//			if post.Price != nil {
//				value = *post.Price
//			}
//			goal.Current += int(math.Round(value))
//			err = s.repository.UpdateGoal(ctx, &goal)
//		}
//	}
//	return err
//}
//
//func (s *Service) GoalType3AddDonationValue(ctx context.Context, donation *Donation) error {
//	if donation.Currency != CurrencyRub {
//		return nil
//	}
//	goals, err := s.repository.GoalsByBlogId(ctx, donation.BlogId)
//	if err != nil {
//		return err
//	}
//	for _, goal := range goals {
//		if goal.Type == "3" {
//			goal.Current += int(math.Round(donation.Value))
//			err = s.repository.UpdateGoal(ctx, &goal)
//		}
//	}
//	return err
//}
