package blogs

import (
	"context"
	"log"
	"time"
)

func (s *Service) StartUserSubscriptionsWorker(ctx context.Context, ticker *time.Ticker) {
	for range ticker.C {

		allUserSubscriptions, err := s.repository.ActiveUserSubscriptionsWithExpiration(ctx)
		if err != nil {
			log.Println("error worker getting active subscriptions info:", err)
			continue
		}

		subscriptionsToUpdate := make([]UserSubscription, len(allUserSubscriptions))
		subscriptionsToUpdateCount := 0
		for i := range allUserSubscriptions {
			if allUserSubscriptions[i].Status == UserSubscriptionStatusCancelled {
				allUserSubscriptions[i].Status = UserSubscriptionStatusExpired
				allUserSubscriptions[i].IsActive = false
				subscriptionsToUpdate[subscriptionsToUpdateCount] = allUserSubscriptions[i]
				subscriptionsToUpdateCount++
			}
		}

		subscriptionsToUpdate = subscriptionsToUpdate[:subscriptionsToUpdateCount]

		err = s.repository.BulkUpdateUserSubscriptionsStatus(ctx, subscriptionsToUpdate)
		if err != nil {
			log.Println("error worker updating subscriptions status:", err)
		}

	}
}

func (r *Repository) ActiveUserSubscriptionsWithExpiration(ctx context.Context) ([]UserSubscription, error) {
	query := `select id, user_id, subscription_id, blog_id, status, is_active, expires_at, created, updated 
			from user_subscriptions
			where is_active = true and expires_at < $1
			order by blog_id, created desc
			`

	rows, err := r.db.Query(ctx, query, time.Now().UTC())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	resultArray := make([]UserSubscription, 0)
	var userSubscription UserSubscription
	for rows.Next() {
		err = rows.Scan(
			&userSubscription.ID,
			&userSubscription.UserId,
			&userSubscription.SubscriptionId,
			&userSubscription.BlogId,
			&userSubscription.Status,
			&userSubscription.IsActive,
			&userSubscription.ExpiresAt,
			&userSubscription.Created,
			&userSubscription.Updated,
		)
		if err != nil {
			return nil, err
		}
		resultArray = append(resultArray, userSubscription)
	}

	return resultArray, nil
}

func (r *Repository) BulkUpdateUserSubscriptionsStatus(ctx context.Context, subscriptions []UserSubscription) error {

	if len(subscriptions) == 0 {
		return nil
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for i := range subscriptions {
		_, err = tx.Exec(ctx, `UPDATE user_subscriptions SET is_active = $1, status = $2 WHERE id = $3`,
			subscriptions[i].IsActive, subscriptions[i].Status, subscriptions[i].ID)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}
