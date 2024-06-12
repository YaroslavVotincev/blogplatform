package blogs

import (
	"context"
	"github.com/google/uuid"
	"log"
	"time"
)

func (s *Service) StartLikesWorker(ctx context.Context, ticker *time.Ticker) {
	for range ticker.C {
		info, err := s.repository.GetPostLikesCountWorkerInfo(ctx)
		if err != nil {
			log.Println("error worker getting likes info:", err)
			continue
		}

		err = s.repository.BulkUpdatePostLikesCount(ctx, info)
		if err != nil {
			log.Println("error worker updating likes:", err)
		}
	}
}

type PostLikesWorkerInfo struct {
	PostId     uuid.UUID
	LikesCount int
}

func (r *Repository) GetPostLikesCountWorkerInfo(ctx context.Context) ([]PostLikesWorkerInfo, error) {

	query := `SELECT post_id, COUNT(CASE WHEN positive = true THEN id  END) AS likes_count
				FROM post_likes
				GROUP BY post_id;`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var resultArray = make([]PostLikesWorkerInfo, 0)
	var item PostLikesWorkerInfo
	for rows.Next() {
		err = rows.Scan(
			&item.PostId,
			&item.LikesCount,
		)
		if err != nil {
			return nil, err
		}
		resultArray = append(resultArray, item)
	}
	return resultArray, nil
}

func (r *Repository) BulkUpdatePostLikesCount(ctx context.Context, likesInfo []PostLikesWorkerInfo) error {

	if len(likesInfo) == 0 {
		return nil
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for i := range likesInfo {
		_, err = tx.Exec(ctx, `UPDATE posts SET likes_count = $1 WHERE id = $2`, likesInfo[i].LikesCount, likesInfo[i].PostId)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}
