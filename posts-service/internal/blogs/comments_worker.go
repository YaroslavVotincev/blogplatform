package blogs

import (
	"context"
	"log"
	"time"
)

func (s *Service) StartCommentsWorker(ctx context.Context, ticker *time.Ticker) {
	for range ticker.C {

		allPosts, err := s.repository.AllPosts(ctx)
		if err != nil {
			log.Println("error worker getting all posts info:", err)
			continue
		}

		for i := range allPosts {
			count, err := s.commentsService.CountPostComments(allPosts[i].ID)
			if err != nil {
				log.Println("error worker getting comments count from service:", err)
			}
			allPosts[i].CommentsCount = count
		}

		err = s.repository.BulkUpdatePostCommentsCount(ctx, allPosts)
		if err != nil {
			log.Println("error worker updating comments count:", err)
		}

	}
}

func (r *Repository) BulkUpdatePostCommentsCount(ctx context.Context, posts []Post) error {

	if len(posts) == 0 {
		return nil
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for i := range posts {
		_, err = tx.Exec(ctx, `UPDATE posts SET comments_count = $1 WHERE id = $2`,
			posts[i].CommentsCount, posts[i].ID)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}
