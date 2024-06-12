package blogs

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/gosimple/slug"
	"net/url"
	"posts-service/internal/billing"
	"posts-service/internal/comments"
	"posts-service/internal/files"
	"posts-service/internal/notifications"
	"posts-service/internal/users"
	configService "posts-service/pkg/config-client"
	"strings"
	"sync"
	"time"
)

type Service struct {
	repository      *Repository
	filesService    *files.Service
	billingService  *billing.Service
	commentsService *comments.Service
	usersService    *users.Service
	notifService    *notifications.Service

	mu *sync.RWMutex

	goalsTicker            *time.Ticker
	likesTicker            *time.Ticker
	commentsTicker         *time.Ticker
	userSubscriptionTicker *time.Ticker
	incomeTicker           *time.Ticker

	mainPageLikesRequirement    int
	mainPageCommentsRequirement int
	mainPageViewsRequirement    int
	mainPageDislikesRequirement int

	donationsRobokassaMinValue float64
	donationsToncoinMinValue   float64
}

func NewService(repository *Repository,
	filesService *files.Service, billingService *billing.Service, commentsService *comments.Service, usersService *users.Service, notifService *notifications.Service,
	mpLikesReq, mpCommentsReq, mpViewsReq, mpDislikesReq int, donatRobokassaMinValue, donatToncoinMinValue float64,
	cfgService *configService.ConfigServiceManager) *Service {

	service := &Service{
		repository:      repository,
		filesService:    filesService,
		billingService:  billingService,
		commentsService: commentsService,
		usersService:    usersService,
		notifService:    notifService,

		mu: &sync.RWMutex{},

		mainPageLikesRequirement:    mpLikesReq,
		mainPageCommentsRequirement: mpCommentsReq,
		mainPageViewsRequirement:    mpViewsReq,
		mainPageDislikesRequirement: mpDislikesReq,

		donationsRobokassaMinValue: donatRobokassaMinValue,
		donationsToncoinMinValue:   donatToncoinMinValue,
	}

	service.goalsTicker = time.NewTicker(1 * time.Minute)
	go service.StartGoalsWorker(context.Background(), service.goalsTicker)

	service.likesTicker = time.NewTicker(1 * time.Minute)
	go service.StartLikesWorker(context.Background(), service.likesTicker)

	service.commentsTicker = time.NewTicker(1 * time.Minute)
	go service.StartCommentsWorker(context.Background(), service.commentsTicker)

	service.userSubscriptionTicker = time.NewTicker(1 * time.Minute)
	go service.StartUserSubscriptionsWorker(context.Background(), service.userSubscriptionTicker)

	service.incomeTicker = time.NewTicker(1 * time.Minute)
	go service.StartBlogIncomeWorker(context.Background(), service.incomeTicker)

	service.SetConfigUpdateHandlers(cfgService)

	return service
}

func (s *Service) StopWorkers() {
	s.goalsTicker.Stop()
	s.likesTicker.Stop()
	s.commentsTicker.Stop()
	s.userSubscriptionTicker.Stop()
	s.incomeTicker.Stop()
}

func (s *Service) BlogById(ctx context.Context, id uuid.UUID) (*Blog, error) {
	return s.repository.BlogById(ctx, id)
}

func (s *Service) BlogByUrl(ctx context.Context, url string) (*Blog, error) {
	return s.repository.BlogByUrl(ctx, url)
}

func (s *Service) NewPersonalBlog(ctx context.Context, userId uuid.UUID) (*Blog, error) {
	id := uuid.New()
	timeNow := time.Now().UTC()
	blog := Blog{
		ID:               id,
		AuthorId:         userId,
		Type:             BlogTypePersonal,
		Url:              "my-new-personal-blog-" + id.String(),
		Title:            "Мой новый персональный блог",
		ShortDescription: "",
		Status:           BlogStatusDraft,
		AcceptDonations:  false,
		Avatar:           nil,
		Cover:            nil,
		Categories:       []string{},
		Created:          timeNow,
		Updated:          timeNow,
	}

	return &blog, s.repository.CreateBlog(ctx, &blog)
}

func (s *Service) NewThematicBlog(ctx context.Context, userId uuid.UUID) (*Blog, error) {
	id := uuid.New()
	timeNow := time.Now().UTC()
	blog := Blog{
		ID:               id,
		AuthorId:         userId,
		Type:             BlogTypeThematic,
		Url:              "my-new-thematic-blog-" + id.String(),
		Title:            "Мой новый тематический блог",
		ShortDescription: "",
		Status:           BlogStatusDraft,
		AcceptDonations:  false,
		Avatar:           nil,
		Cover:            nil,
		Categories:       []string{},
		Created:          timeNow,
		Updated:          timeNow,
	}

	return &blog, s.repository.CreateBlog(ctx, &blog)
}

func (s *Service) RedirectUrlToFile(filename string) (string, error) {
	path, err := url.JoinPath(s.filesService.GetFileEndpointUrl(), filename)
	if err != nil {
		return "", err
	}
	return path, nil
}

func (s *Service) SetBlogAvatar(ctx context.Context, blog *Blog, bytes []byte) error {
	if blog.Avatar == nil {
		id := uuid.New().String()
		blog.Avatar = &id
		err := s.repository.UpdateBlog(ctx, blog)
		if err != nil {
			return err
		}
	}
	go s.filesService.SendFile(*blog.Avatar, bytes)
	return nil
}

func (s *Service) SetBlogCover(ctx context.Context, blog *Blog, bytes []byte) error {
	if blog.Cover == nil {
		id := uuid.New().String()
		blog.Cover = &id
		blog.Updated = time.Now().UTC()
		err := s.repository.UpdateBlog(ctx, blog)
		if err != nil {
			return err
		}
	}
	go s.filesService.SendFile(*blog.Cover, bytes)
	return nil
}

func (s *Service) PostById(ctx context.Context, id uuid.UUID) (*Post, error) {
	return s.repository.PostById(ctx, id)
}

func (s *Service) CreatePostToBlog(ctx context.Context, blogId uuid.UUID) (*Post, error) {
	timeNow := time.Now().UTC()
	id := uuid.New()
	post := Post{
		ID:               id,
		BlogId:           blogId,
		Title:            "Моя новая публикация",
		Url:              fmt.Sprintf("my-new-post-%s", id.String()),
		ShortDescription: "",
		TagsString:       "",
		Status:           PostStatusDraft,
		Cover:            nil,
		AccessMode:       "1",
		Price:            nil,
		SubscriptionId:   nil,
		Created:          timeNow,
		Updated:          timeNow,
	}
	return &post, s.repository.CreatePost(ctx, &post)
}

func (s *Service) SetTagsToPost(ctx context.Context, postId uuid.UUID, tagsString string) error {
	tags := strings.Split(tagsString, ",")
	for i, tag := range tags {
		tags[i] = slug.Make(tag)
	}
	return s.repository.SetTagsToPost(ctx, postId, unique(tags))
}

func (s *Service) SetPostCover(ctx context.Context, post *Post, bytes []byte) error {
	if post.Cover == nil {
		id := uuid.New().String()
		post.Cover = &id
		post.Updated = time.Now().UTC()
		err := s.repository.UpdatePost(ctx, post)
		if err != nil {
			return err
		}
	}
	go s.filesService.SendFile(*post.Cover, bytes)
	return nil
}

func (s *Service) ContentById(ctx context.Context, id uuid.UUID) (*Content, error) {
	content, err := s.repository.ContentById(ctx, id)
	if err != nil {
		return nil, err
	}
	if content == nil {
		content, err = s.CreateContentToId(ctx, id)
		if err != nil {
			return nil, err
		}
	}
	return content, nil
}

func (s *Service) CreateContentToId(ctx context.Context, id uuid.UUID) (*Content, error) {
	timeNow := time.Now().UTC()
	content := Content{
		ID:       id,
		DataJson: defaultContentJson,
		DataHtml: defaultContentHtml,
		Created:  timeNow,
		Updated:  timeNow,
	}
	return &content, s.repository.CreateContent(ctx, &content)
}

func (s *Service) CreateFileToContent(ctx context.Context, content *Content, bytes []byte, fileType string) (*ContentFile, error) {
	id := uuid.New()
	timeNow := time.Now().UTC()
	contentFile := ContentFile{
		ID:        id,
		ContentId: content.ID,
		Type:      fileType,
		Size:      len(bytes),
		Created:   timeNow,
		Updated:   timeNow,
	}

	err := s.repository.CreateContentFile(ctx, &contentFile)
	if err != nil {
		return nil, err
	}
	go s.filesService.SendFile(id.String(), bytes)
	return &contentFile, nil
}

func (s *Service) DeleteContentFile(ctx context.Context, contentFile *ContentFile) error {
	go s.filesService.DeleteFile(contentFile.ID.String())
	return s.repository.DeleteContentFile(ctx, contentFile)
}

func (s *Service) SetSubscriptionCover(ctx context.Context, subscription *Subscription, bytes []byte) error {
	if subscription.Cover == nil {
		id := uuid.New().String()
		subscription.Cover = &id
		subscription.Updated = time.Now().UTC()
		err := s.repository.UpdateSubscription(ctx, subscription)
		if err != nil {
			return err
		}
	}
	go s.filesService.SendFile(*subscription.Cover, bytes)
	return nil
}

func (s *Service) GetSubscriptionRobokassaPaymentLink(ctx context.Context, subscription *Subscription, userId uuid.UUID) (string, error) {
	description := fmt.Sprintf("Оплата подписки \"%s\" (ID: %s)", subscription.Title, subscription.ID)
	return s.billingService.RobokassaPaymentLink(subscription.ID, userId, subscription.PriceRub, PaymentItemTypeSubscription, description)
}

func (s *Service) GetPostRobokassaPaymentLink(ctx context.Context, post *Post, userId uuid.UUID) (string, error) {
	description := fmt.Sprintf("Оплата публикации \"%s\" (ID: %s)", post.Title, post.ID)
	return s.billingService.RobokassaPaymentLink(post.ID, userId, *post.Price, PaymentItemTypePost, description)
}

func (s *Service) GetDonationRobokassaPaymentLink(blog *Blog, donation *Donation, userId uuid.UUID) (string, error) {
	description := fmt.Sprintf("Оплата пожертвования (ID: %s) для развития блога \"%s\" (ID: %s)",
		donation.ID.String(), blog.Title, blog.ID.String())
	return s.billingService.RobokassaPaymentLink(donation.ID, userId, donation.Value, PaymentItemTypeDonation, description)
}

func (s *Service) PostLikesInfo(ctx context.Context, postId uuid.UUID, userId uuid.UUID) (*PostLikesInfoResponse, error) {
	likes, dislikes, err := s.repository.CountPostLikes(ctx, postId)
	if err != nil {
		return nil, err
	}
	postLike, err := s.repository.PostLikeByPostIdAndUserId(ctx, postId, userId)
	if err != nil {
		return nil, err
	}
	return &PostLikesInfoResponse{
		Likes:     likes,
		Dislikes:  dislikes,
		MyLike:    postLike != nil && postLike.Positive,
		MyDislike: postLike != nil && !postLike.Positive,
	}, nil
}

func (s *Service) LikePost(ctx context.Context, postId uuid.UUID, userId uuid.UUID) error {
	postLike, err := s.repository.PostLikeByPostIdAndUserId(ctx, postId, userId)
	if err != nil {
		return err
	}
	if postLike == nil {
		return s.repository.CreatePostLike(ctx, &PostLike{
			ID:       uuid.New(),
			PostId:   postId,
			UserId:   userId,
			Positive: true,
			Created:  time.Now().UTC(),
		})
	} else {
		if postLike.Positive {
			return nil
		}
		postLike.Positive = true
		postLike.Created = time.Now().UTC()
		return s.repository.UpdatePostLike(ctx, postLike)
	}
}

func (s *Service) DislikePost(ctx context.Context, postId uuid.UUID, userId uuid.UUID) error {
	postLike, err := s.repository.PostLikeByPostIdAndUserId(ctx, postId, userId)
	if err != nil {
		return err
	}
	if postLike == nil {
		return s.repository.CreatePostLike(ctx, &PostLike{
			ID:       uuid.New(),
			PostId:   postId,
			UserId:   userId,
			Positive: false,
			Created:  time.Now().UTC(),
		})
	} else {
		if !postLike.Positive {
			return nil
		}
		postLike.Positive = false
		postLike.Created = time.Now().UTC()
		return s.repository.UpdatePostLike(ctx, postLike)
	}
}

func (s *Service) UnsetLike(ctx context.Context, postId uuid.UUID, userId uuid.UUID) error {
	postLike, err := s.repository.PostLikeByPostIdAndUserId(ctx, postId, userId)
	if err != nil {
		return err
	}
	if postLike == nil {
		return nil
	}
	return s.repository.DeletePostLike(ctx, postLike)
}

func (s *Service) AddPostViewWithUser(ctx context.Context, postId uuid.UUID, userId uuid.UUID) error {
	if userId == uuid.Nil {
		return nil
	}

	postView, err := s.repository.PostViewByPostIdAndUserId(ctx, postId, userId)
	if err != nil {
		return err
	}
	if postView == nil {
		return s.repository.CreatePostView(ctx, &PostView{
			ID:          uuid.New(),
			PostId:      postId,
			UserId:      &userId,
			Fingerprint: nil,
			Created:     time.Now().UTC(),
		})
	} else {
		return nil
	}
}

func (s *Service) AddPostViewWithFingerprint(ctx context.Context, postId uuid.UUID, fingerprint string) error {

	postView, err := s.repository.PostViewByPostIdAndFingerprint(ctx, postId, fingerprint)
	if err != nil {
		return err
	}
	if postView == nil {
		return s.repository.CreatePostView(ctx, &PostView{
			ID:          uuid.New(),
			PostId:      postId,
			UserId:      nil,
			Fingerprint: &fingerprint,
			Created:     time.Now().UTC(),
		})
	} else {
		return nil
	}
}

func (s *Service) CheckUserContentAccess(ctx context.Context, post *Post, userId uuid.UUID) (bool, error) {
	blog, err := s.BlogById(ctx, post.BlogId)
	if err != nil {
		return false, fmt.Errorf("failed to get blog by id: %w", err)
	}
	if blog == nil {
		return false, fmt.Errorf("blog by id doesn't exists")
	}
	if blog.AuthorId == userId {
		return true, nil
	}

	blogSubscriptions, err := s.repository.SubscriptionsByBlogId(ctx, post.BlogId)
	if err != nil {
		return false, fmt.Errorf("failed to get subscriptions by blog id: %w", err)
	}

	switch post.AccessMode {

	case "1":
		return true, nil

	case "2":
		haveFreeSubscription := false
		for _, sub := range blogSubscriptions {
			if sub.IsFree {
				haveFreeSubscription = true
				break
			}
		}
		if !haveFreeSubscription {
			return true, nil
		}
		follow, err := s.repository.UserFollowByUserIdAndBlogId(ctx, userId, post.BlogId)
		if err != nil {
			return false, fmt.Errorf("failed to get follow by user id and blog id: %w", err)
		}
		return follow != nil, nil

	case "3":
		firstPaidSubscriptionIdx := 0
		havePaidSubscription := false
		for i, sub := range blogSubscriptions {
			if !sub.IsFree {
				firstPaidSubscriptionIdx = i
				havePaidSubscription = true
				break
			}
		}
		if !havePaidSubscription {
			return true, nil
		}
		if post.SubscriptionId == nil {
			tempId := blogSubscriptions[firstPaidSubscriptionIdx].ID
			post.SubscriptionId = &tempId
			err = s.repository.UpdatePost(ctx, post)
			if err != nil {
				return false, fmt.Errorf("failed to update post changing nil subscription id to first paid subscription id: %w", err)
			}
		}

		var reqSubId = 0
		for i, sub := range blogSubscriptions {
			if sub.ID == *post.SubscriptionId {
				reqSubId = i
				break
			}
		}
		var requiredSubscriptions = blogSubscriptions[reqSubId:]

		userSubs, err := s.repository.UserSubscriptionsByUserIdAndBlogId(ctx, userId, post.BlogId)
		if err != nil {
			return false, fmt.Errorf("failed to get subscriptions by user id and blog id: %w", err)
		}

		for _, sub := range requiredSubscriptions {
			for _, userSub := range userSubs {
				if sub.ID == userSub.SubscriptionId && userSub.IsActive {
					return true, nil
				}
			}
		}
		return false, nil

	case "4":
		userPaidAccess, err := s.repository.PostPaidAccessByPostIdAndUserId(ctx, post.ID, userId)
		if err != nil {
			return false, fmt.Errorf("failed to get paid access by post id and user id: %w", err)
		}
		return userPaidAccess != nil, nil

	default:
		return false, fmt.Errorf("unknown access mode: %s", post.AccessMode)
	}
}
