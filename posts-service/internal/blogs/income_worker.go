package blogs

import (
	"context"
	"github.com/google/uuid"
	"time"
)

func (s *Service) StartBlogIncomeWorker(ctx context.Context, ticker *time.Ticker) {
	for range ticker.C {
		blogIncomes, _ := s.repository.UnsentBlogIncomes(ctx)
		blogsIdsStr := make([]string, 0)
		for _, blogIncome := range blogIncomes {
			blogsIdsStr = append(blogsIdsStr, blogIncome.BlogId.String())
		}
		blogs, _ := s.repository.BlogsByIdList(ctx, blogsIdsStr)
		blogsMap := make(map[uuid.UUID]Blog, len(blogs))
		usersMap := make(map[uuid.UUID]float64)
		for _, blog := range blogs {
			blogsMap[blog.ID] = blog
			usersMap[blog.AuthorId] = 0
		}
		for _, blogIncome := range blogIncomes {
			blog := blogsMap[blogIncome.BlogId]
			userAmount := usersMap[blog.AuthorId]
			usersMap[blog.AuthorId] = userAmount + blogIncome.Value
		}
		for userId, userAmount := range usersMap {
			_ = s.usersService.AddRubToBalanceOfUser(userId, userAmount)
		}

		_ = s.repository.SetBlogIncomesSent(ctx, blogIncomes)
	}
}

func (r *Repository) UnsentBlogIncomes(ctx context.Context) ([]BlogIncome, error) {
	query := `select id, blog_id, user_id, value, currency, item_id, item_type, sent_to_user_wallet, created 
			from blog_incomes
		 	where sent_to_user_wallet = false
		 	order by created desc`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var item BlogIncome
	var blogIncomes = make([]BlogIncome, 0)
	for rows.Next() {
		err := rows.Scan(
			&item.ID,
			&item.BlogId,
			&item.UserId,
			&item.Value,
			&item.Currency,
			&item.ItemId,
			&item.ItemType,
			&item.SentToUserWallet,
			&item.Created,
		)
		if err != nil {
			return nil, err
		}
		blogIncomes = append(blogIncomes, item)
	}
	return blogIncomes, nil
}

func (r *Repository) SetBlogIncomesSent(ctx context.Context, blogIncomes []BlogIncome) error {
	ids := make([]uuid.UUID, len(blogIncomes))
	for i := range blogIncomes {
		ids[i] = blogIncomes[i].ID
	}
	query := `update blog_incomes set sent_to_user_wallet = true where id = ANY($1)`
	_, err := r.db.Exec(ctx, query, ids)
	return err
}
