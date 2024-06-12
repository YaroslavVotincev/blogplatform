package blogs

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"strings"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) AllBlogs(ctx context.Context) ([]Blog, error) {
	query := `select id, author_id, type, url, title, short_description, 
       status, accept_donations, avatar, cover, c.categories,
       created, updated
	from blogs
	left JOIN  (select bc.blog_id AS id, array_agg(c.code) as categories
   				from blog_categories bc
   				join categories c on c.code = bc.category
   				group by bc.blog_id
   	) c USING (id)
	order by created desc`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	resultArray := make([]Blog, 0)
	var item Blog
	for rows.Next() {
		err = rows.Scan(
			&item.ID,
			&item.AuthorId,
			&item.Type,
			&item.Url,
			&item.Title,
			&item.ShortDescription,
			&item.Status,
			&item.AcceptDonations,
			&item.Avatar,
			&item.Cover,
			&item.Categories,
			&item.Created,
			&item.Updated,
		)
		if err != nil {
			return nil, err
		}
		if item.Categories == nil {
			item.Categories = []string{}
		}
		resultArray = append(resultArray, item)
	}
	return resultArray, nil
}

func (r *Repository) BlogsByIdList(ctx context.Context, ids []string) ([]Blog, error) {
	if len(ids) == 0 {
		return make([]Blog, 0), nil
	}
	query := `select id, author_id, type, url, title, short_description, 
       status, accept_donations, avatar, cover, c.categories,
       created, updated
	from blogs
	left join  (select bc.blog_id AS id, array_agg(c.code) as categories
   				from blog_categories bc
   				join categories c on c.code = bc.category
   				group by bc.blog_id
   	) c using (id) 
	where id = Any ($1)
	order by created desc`
	rows, err := r.db.Query(ctx, query, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	resultArray := make([]Blog, 0)
	var item Blog
	for rows.Next() {
		err = rows.Scan(
			&item.ID,
			&item.AuthorId,
			&item.Type,
			&item.Url,
			&item.Title,
			&item.ShortDescription,
			&item.Status,
			&item.AcceptDonations,
			&item.Avatar,
			&item.Cover,
			&item.Categories,
			&item.Created,
			&item.Updated,
		)
		if err != nil {
			return nil, err
		}
		if item.Categories == nil {
			item.Categories = []string{}
		}
		resultArray = append(resultArray, item)
	}
	return resultArray, nil
}

func (r *Repository) BlogsByUserId(ctx context.Context, userId uuid.UUID) ([]Blog, error) {
	query := `select id, author_id, type, url, title, short_description, 
       status, accept_donations, avatar, cover, c.categories,
       created, updated
	from blogs
	left JOIN  (select bc.blog_id AS id, array_agg(c.code) as categories
   				from blog_categories bc
   				join categories c on c.code = bc.category
   				group by bc.blog_id
   	) c USING (id)
	where author_id = $1
	order by created desc`

	rows, err := r.db.Query(ctx, query, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	resultArray := make([]Blog, 0)
	var item Blog
	for rows.Next() {
		err = rows.Scan(
			&item.ID,
			&item.AuthorId,
			&item.Type,
			&item.Url,
			&item.Title,
			&item.ShortDescription,
			&item.Status,
			&item.AcceptDonations,
			&item.Avatar,
			&item.Cover,
			&item.Categories,
			&item.Created,
			&item.Updated,
		)
		if err != nil {
			return nil, err
		}
		if item.Categories == nil {
			item.Categories = []string{}
		}
		resultArray = append(resultArray, item)
	}
	return resultArray, nil
}

func (r *Repository) BlogById(ctx context.Context, id uuid.UUID) (*Blog, error) {
	query := `select id, author_id, type, url, title, short_description, 
       status, accept_donations, avatar, cover, c.categories,
       created, updated
	from blogs, LATERAL (  
   			SELECT ARRAY (
      			SELECT c.code
      			FROM   blog_categories bc
      			JOIN   categories c ON c.code = bc.category
      			WHERE  bc.blog_id = id
      	) AS categories
   	) c
	where id = $1
	order by created desc`
	var blog Blog
	err := r.db.QueryRow(ctx, query, id).Scan(
		&blog.ID,
		&blog.AuthorId,
		&blog.Type,
		&blog.Url,
		&blog.Title,
		&blog.ShortDescription,
		&blog.Status,
		&blog.AcceptDonations,
		&blog.Avatar,
		&blog.Cover,
		&blog.Categories,
		&blog.Created,
		&blog.Updated,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &blog, nil
}

func (r *Repository) BlogByUrl(ctx context.Context, url string) (*Blog, error) {
	query := `select id, author_id, type, url, title, short_description, 
       status, accept_donations, avatar, cover, c.categories,
       created, updated
	from blogs, LATERAL (  
   			SELECT ARRAY (
      			SELECT c.code
      			FROM   blog_categories bc
      			JOIN   categories c ON c.code = bc.category
      			WHERE  bc.blog_id = id
      	) AS categories
   	) c
	where url = $1
	order by created desc`
	var blog Blog
	err := r.db.QueryRow(ctx, query, url).Scan(
		&blog.ID,
		&blog.AuthorId,
		&blog.Type,
		&blog.Url,
		&blog.Title,
		&blog.ShortDescription,
		&blog.Status,
		&blog.AcceptDonations,
		&blog.Avatar,
		&blog.Cover,
		&blog.Categories,
		&blog.Created,
		&blog.Updated,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &blog, nil
}

func (r *Repository) CreateBlog(ctx context.Context, blog *Blog) error {
	query := `insert into blogs
	(id, author_id, type, url, title, short_description, status, accept_donations, avatar, cover, created, updated)
	values
	($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`
	_, err := r.db.Exec(ctx, query,
		blog.ID,
		blog.AuthorId,
		blog.Type,
		blog.Url,
		blog.Title,
		blog.ShortDescription,
		blog.Status,
		blog.AcceptDonations,
		blog.Avatar,
		blog.Cover,
		blog.Created,
		blog.Updated,
	)
	return err
}

func (r *Repository) UpdateBlog(ctx context.Context, blog *Blog) error {
	query := `update blogs set
	author_id = $2,
	type = $3,
	url = $4,
	title = $5,
	short_description = $6,
	status = $7,
	accept_donations = $8,
	avatar = $9,
	cover = $10,
	created = $11,
	updated = $12
	where id = $1`
	_, err := r.db.Exec(ctx, query,
		&blog.ID,
		&blog.AuthorId,
		&blog.Type,
		&blog.Url,
		&blog.Title,
		&blog.ShortDescription,
		&blog.Status,
		&blog.AcceptDonations,
		&blog.Avatar,
		&blog.Cover,
		&blog.Created,
		&blog.Updated,
	)
	return err
}

func (r *Repository) AllCategories(ctx context.Context) ([]Category, error) {
	query := `SELECT c.code, c.name, c.created, c.updated, COUNT(p.id) AS post_count
		FROM categories c
		LEFT JOIN blog_categories bc ON c.code = bc.category
		LEFT JOIN posts p ON bc.blog_id = p.blog_id
		GROUP BY c.code, c.name, c.created, c.updated
		ORDER BY COUNT(p.id) DESC;
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	resultArray := make([]Category, 0)
	var num int
	var item Category
	for rows.Next() {
		err = rows.Scan(
			&item.Code,
			&item.Name,
			&item.Created,
			&item.Updated,
			&num,
		)
		if err != nil {
			return nil, err
		}
		resultArray = append(resultArray, item)
	}
	return resultArray, nil
}

func (r *Repository) CategoriesByUserPreferences(ctx context.Context, userId uuid.UUID) ([]Category, error) {
	query := `SELECT c.code, c.name, c.created, c.updated AS post_count
		FROM user_categories_preferences ucp
		LEFT JOIN categories c ON ucp.category = c.code
		where user_id = $1
	`

	rows, err := r.db.Query(ctx, query, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	resultArray := make([]Category, 0)
	var item Category
	for rows.Next() {
		err = rows.Scan(
			&item.Code,
			&item.Name,
			&item.Created,
			&item.Updated,
		)
		if err != nil {
			return nil, err
		}
		resultArray = append(resultArray, item)
	}
	return resultArray, nil
}

func (r *Repository) SetBlogCategories(ctx context.Context, blogId uuid.UUID, categories []string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	query := `delete from blog_categories where blog_id = $1`
	_, err = tx.Exec(ctx, query, blogId)
	if err != nil {
		return err
	}

	if len(categories) == 0 {
		return tx.Commit(ctx)
	}

	query = `insert into blog_categories (blog_id, category) values `
	values := make([]any, 2*len(categories))
	for i := range categories {
		query += fmt.Sprintf(`($%d, $%d)`, i*2+1, i*2+2)
		values[i*2] = blogId
		values[i*2+1] = categories[i]
		if i != len(categories)-1 {
			query += ","
		}
	}
	_, err = tx.Exec(ctx, query, values...)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *Repository) SetUserCategoriesPreferences(ctx context.Context, userId uuid.UUID, categories []string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	query := `delete from user_categories_preferences where user_id = $1`
	_, err = tx.Exec(ctx, query, userId)
	if err != nil {
		return err
	}

	if len(categories) == 0 {
		return tx.Commit(ctx)
	}

	query = `insert into user_categories_preferences (user_id, category) values `
	values := make([]any, 2*len(categories))
	for i := range categories {
		query += fmt.Sprintf(`($%d, $%d)`, i*2+1, i*2+2)
		values[i*2] = userId
		values[i*2+1] = categories[i]
		if i != len(categories)-1 {
			query += ","
		}
	}
	_, err = tx.Exec(ctx, query, values...)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *Repository) AllPosts(ctx context.Context) ([]Post, error) {

	query := `select id, blog_id, title, url, short_description, tags_string, status, cover, 
       access_mode, price, subscription_id, likes_count, comments_count, created, updated
	from posts
	order by created desc`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	resultArray := make([]Post, 0)
	var post Post
	for rows.Next() {
		err = rows.Scan(
			&post.ID,
			&post.BlogId,
			&post.Title,
			&post.Url,
			&post.ShortDescription,
			&post.TagsString,
			&post.Status,
			&post.Cover,
			&post.AccessMode,
			&post.Price,
			&post.SubscriptionId,
			&post.LikesCount,
			&post.CommentsCount,
			&post.Created,
			&post.Updated,
		)
		if err != nil {
			return nil, err
		}
		resultArray = append(resultArray, post)
	}
	return resultArray, nil

}

func (r *Repository) PostsByCategories(ctx context.Context, categoryCodes []string) ([]Post, error) {
	query := `SELECT p.id, p.blog_id, p.title, p.url, p.short_description, p.tags_string, p.status, p.cover,
              p.access_mode, p.price, p.subscription_id, p.likes_count, p.comments_count, p.created, p.updated
              FROM posts p
              JOIN blogs b ON p.blog_id = b.id
              JOIN blog_categories bc ON b.id = bc.blog_id
              WHERE bc.category = Any ($1)
              ORDER BY p.created DESC;`

	rows, err := r.db.Query(ctx, query, categoryCodes)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts = make([]Post, 0)
	for rows.Next() {
		var post Post
		err = rows.Scan(
			&post.ID,
			&post.BlogId,
			&post.Title,
			&post.Url,
			&post.ShortDescription,
			&post.TagsString,
			&post.Status,
			&post.Cover,
			&post.AccessMode,
			&post.Price,
			&post.SubscriptionId,
			&post.LikesCount,
			&post.CommentsCount,
			&post.Created,
			&post.Updated,
		)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}
	return posts, nil
}

func (r *Repository) PostById(ctx context.Context, id uuid.UUID) (*Post, error) {
	query := `select id, blog_id, title, url, short_description, tags_string, status, cover, 
       access_mode, price, subscription_id, likes_count, comments_count, created, updated
	from posts
	where id = $1`
	var post Post
	err := r.db.QueryRow(ctx, query, id).Scan(
		&post.ID,
		&post.BlogId,
		&post.Title,
		&post.Url,
		&post.ShortDescription,
		&post.TagsString,
		&post.Status,
		&post.Cover,
		&post.AccessMode,
		&post.Price,
		&post.SubscriptionId,
		&post.LikesCount,
		&post.CommentsCount,
		&post.Created,
		&post.Updated,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &post, nil
}

func (r *Repository) PostsByBlogId(ctx context.Context, blogId uuid.UUID) ([]Post, error) {

	query := `select id, blog_id, title, url, short_description, tags_string, status, cover, 
       access_mode, price, subscription_id, likes_count, comments_count, created, updated
	from posts
	where blog_id = $1
	order by created desc`

	rows, err := r.db.Query(ctx, query, blogId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	resultArray := make([]Post, 0)
	var post Post
	for rows.Next() {
		err = rows.Scan(
			&post.ID,
			&post.BlogId,
			&post.Title,
			&post.Url,
			&post.ShortDescription,
			&post.TagsString,
			&post.Status,
			&post.Cover,
			&post.AccessMode,
			&post.Price,
			&post.SubscriptionId,
			&post.LikesCount,
			&post.CommentsCount,
			&post.Created,
			&post.Updated,
		)
		if err != nil {
			return nil, err
		}
		resultArray = append(resultArray, post)
	}
	return resultArray, nil
}
func (r *Repository) PostsByBlogIdList(ctx context.Context, ids []string) ([]Post, error) {
	if len(ids) == 0 {
		return make([]Post, 0), nil
	}
	query := `select id, blog_id, title, url, short_description, tags_string, status, cover, 
       access_mode, price, subscription_id, likes_count, comments_count, created, updated
	from posts`
	placeholders := make([]string, len(ids))
	idInterfaceSlice := make([]interface{}, len(ids))
	for i := range ids {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		idInterfaceSlice[i] = ids[i]
	}
	query += fmt.Sprintf(" where blog_id in (%s)", strings.Join(placeholders, ","))
	query += " order by created desc"

	rows, err := r.db.Query(ctx, query, idInterfaceSlice...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	resultArray := make([]Post, 0)
	var post Post
	for rows.Next() {
		err = rows.Scan(
			&post.ID,
			&post.BlogId,
			&post.Title,
			&post.Url,
			&post.ShortDescription,
			&post.TagsString,
			&post.Status,
			&post.Cover,
			&post.AccessMode,
			&post.Price,
			&post.SubscriptionId,
			&post.LikesCount,
			&post.CommentsCount,
			&post.Created,
			&post.Updated,
		)
		if err != nil {
			return nil, err
		}
		resultArray = append(resultArray, post)
	}
	return resultArray, nil
}

func (r *Repository) PostsByBlogIdAndUrl(ctx context.Context, blogId uuid.UUID, url string) (*Post, error) {

	query := `select id, blog_id, title, url, short_description, tags_string, status, cover, 
       access_mode, price, subscription_id, likes_count, comments_count, created, updated
	from posts
	where blog_id = $1 and url = $2
	order by created desc`

	var post Post
	err := r.db.QueryRow(ctx, query, blogId, url).Scan(
		&post.ID,
		&post.BlogId,
		&post.Title,
		&post.Url,
		&post.ShortDescription,
		&post.TagsString,
		&post.Status,
		&post.Cover,
		&post.AccessMode,
		&post.Price,
		&post.SubscriptionId,
		&post.LikesCount,
		&post.CommentsCount,
		&post.Created,
		&post.Updated,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &post, nil
}

func (r *Repository) CreatePost(ctx context.Context, post *Post) error {
	query := `insert into posts
	(id, blog_id, title, url, short_description, tags_string, status, cover, 
	 access_mode, price, subscription_id, created, updated)
	values
	($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`
	_, err := r.db.Exec(ctx, query,
		post.ID,
		post.BlogId,
		post.Title,
		post.Url,
		post.ShortDescription,
		post.TagsString,
		post.Status,
		post.Cover,
		post.AccessMode,
		post.Price,
		post.SubscriptionId,
		post.Created,
		post.Updated,
	)
	return err
}

func (r *Repository) UpdatePost(ctx context.Context, post *Post) error {
	query := `update posts set
	blog_id = $2,
	title = $3,
	url = $4,
	short_description = $5,
	tags_string = $6,
	status = $7,
	cover = $8,
	access_mode = $9,
	price = $10,
	subscription_id = $11,
	created = $12,
	updated = $13
	where id = $1`
	_, err := r.db.Exec(ctx, query,
		post.ID,
		post.BlogId,
		post.Title,
		post.Url,
		post.ShortDescription,
		post.TagsString,
		post.Status,
		post.Cover,
		post.AccessMode,
		post.Price,
		post.SubscriptionId,
		post.Created,
		post.Updated,
	)
	return err
}

func (r *Repository) SetTagsToPost(ctx context.Context, postId uuid.UUID, tags []string) error {
	if len(tags) == 0 {
		return nil
	}
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	query := `delete from posts_tags where post_id = $1`
	_, err = tx.Exec(ctx, query, postId)
	if err != nil {
		return err
	}
	query = `select id, slug from tags`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return err
	}
	defer rows.Close()
	slugMap := make(map[string]uuid.UUID)
	var id uuid.UUID
	var slug string
	for rows.Next() {
		err = rows.Scan(&id, &slug)
		if err != nil {
			return err
		}
		slugMap[slug] = id
	}
	query1 := `insert into tags (id, slug) values `
	counter1 := 0
	values1 := make([]any, 0)
	query2 := `insert into posts_tags (post_id, tag_id) values `
	counter2 := 0
	values2 := make([]any, 0)
	for _, tag := range tags {
		if _, ok := slugMap[tag]; !ok {
			id := uuid.New()
			query1 += fmt.Sprintf(`($%d, $%d),`, counter1+1, counter1+2)
			values1 = append(values1, id, tag)
			counter1 += 2
			slugMap[tag] = id
		}

		query2 += fmt.Sprintf(`($%d, $%d),`, counter2+1, counter2+2)
		values2 = append(values2, postId, slugMap[tag])
		counter2 += 2
	}
	if counter1 > 0 {
		_, err = tx.Exec(ctx, query1[:len(query1)-1], values1...)
		if err != nil {
			return err
		}
	}
	_, err = tx.Exec(ctx, query2[:len(query2)-1], values2...)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *Repository) ContentById(ctx context.Context, id uuid.UUID) (*Content, error) {

	query := `select id, data_json, data_html, created, updated from contents where id = $1`
	var content Content
	err := r.db.QueryRow(ctx, query, id).Scan(
		&content.ID,
		&content.DataJson,
		&content.DataHtml,
		&content.Created,
		&content.Updated,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &content, nil
}

func (r *Repository) CreateContent(ctx context.Context, content *Content) error {
	query := `insert into contents
	(id, data_json, data_html, created, updated)
	values
	($1, $2, $3, $4, $5)`
	_, err := r.db.Exec(ctx, query,
		content.ID,
		content.DataJson,
		content.DataHtml,
		content.Created,
		content.Updated,
	)
	return err
}

func (r *Repository) UpdateContent(ctx context.Context, content *Content) error {
	query := `update contents set
	data_json = $2,
	data_html = $3,
	created = $4,
	updated = $5
	where id = $1`
	_, err := r.db.Exec(ctx, query,
		&content.ID,
		&content.DataJson,
		&content.DataHtml,
		&content.Created,
		&content.Updated,
	)
	return err
}

func (r *Repository) ContentFileById(ctx context.Context, id uuid.UUID) (*ContentFile, error) {
	query := `select id, content_id, type, size, created, updated from content_files where id = $1`
	var contentFile ContentFile
	err := r.db.QueryRow(ctx, query, id).Scan(
		&contentFile.ID,
		&contentFile.ContentId,
		&contentFile.Type,
		&contentFile.Size,
		&contentFile.Created,
		&contentFile.Updated,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &contentFile, nil
}

func (r *Repository) CreateContentFile(ctx context.Context, contentFile *ContentFile) error {

	query := `insert into content_files
	(id, content_id, type, size, created, updated)
	values
	($1, $2, $3, $4, $5, $6)`
	_, err := r.db.Exec(ctx, query,
		contentFile.ID,
		contentFile.ContentId,
		contentFile.Type,
		contentFile.Size,
		contentFile.Created,
		contentFile.Updated,
	)
	return err
}

func (r *Repository) DeleteContentFile(ctx context.Context, contentFile *ContentFile) error {
	query := `delete from content_files where id = $1`
	_, err := r.db.Exec(ctx, query, contentFile.ID)
	return err
}

func (r *Repository) SubscriptionsByIdList(ctx context.Context, ids []string) ([]Subscription, error) {
	if len(ids) == 0 {
		return make([]Subscription, 0), nil
	}
	query := `select id, blog_id, title, short_description, cover, is_free, price_rub, is_active, created, updated 
			from subscriptions
			where id = any($1)
			order by price_rub, created, is_active desc
			`
	rows, err := r.db.Query(ctx, query, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	resultArray := make([]Subscription, 0)
	var subscription Subscription
	for rows.Next() {
		err = rows.Scan(
			&subscription.ID,
			&subscription.BlogId,
			&subscription.Title,
			&subscription.ShortDescription,
			&subscription.Cover,
			&subscription.IsFree,
			&subscription.PriceRub,
			&subscription.IsActive,
			&subscription.Created,
			&subscription.Updated,
		)
		if err != nil {
			return nil, err
		}
		resultArray = append(resultArray, subscription)
	}

	return resultArray, nil
}

func (r *Repository) SubscriptionsByBlogId(ctx context.Context, blogId uuid.UUID) ([]Subscription, error) {
	query := `select id, blog_id, title, short_description, cover, is_free, price_rub, is_active, created, updated 
			from subscriptions where blog_id = $1
			order by price_rub
			`
	rows, err := r.db.Query(ctx, query, blogId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	resultArray := make([]Subscription, 0)
	var subscription Subscription
	for rows.Next() {
		err = rows.Scan(
			&subscription.ID,
			&subscription.BlogId,
			&subscription.Title,
			&subscription.ShortDescription,
			&subscription.Cover,
			&subscription.IsFree,
			&subscription.PriceRub,
			&subscription.IsActive,
			&subscription.Created,
			&subscription.Updated,
		)
		if err != nil {
			return nil, err
		}
		resultArray = append(resultArray, subscription)
	}

	return resultArray, nil
}

func (r *Repository) SubscriptionById(ctx context.Context, id uuid.UUID) (*Subscription, error) {

	query := `select id, blog_id, title, short_description, cover, is_free, price_rub, is_active, created, updated 
			from subscriptions where id = $1`

	var subscription Subscription
	err := r.db.QueryRow(ctx, query, id).Scan(
		&subscription.ID,
		&subscription.BlogId,
		&subscription.Title,
		&subscription.ShortDescription,
		&subscription.Cover,
		&subscription.IsFree,
		&subscription.PriceRub,
		&subscription.IsActive,
		&subscription.Created,
		&subscription.Updated,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &subscription, nil
}

func (r *Repository) CreateSubscription(ctx context.Context, subscription *Subscription) error {
	query := `insert into subscriptions
	(id, blog_id, title, short_description, cover, is_free, price_rub, is_active, created, updated)
	values
	($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
	_, err := r.db.Exec(ctx, query,
		subscription.ID,
		subscription.BlogId,
		subscription.Title,
		subscription.ShortDescription,
		subscription.Cover,
		subscription.IsFree,
		subscription.PriceRub,
		subscription.IsActive,
		subscription.Created,
		subscription.Updated,
	)
	return err
}

func (r *Repository) UpdateSubscription(ctx context.Context, subscription *Subscription) error {

	query := `update subscriptions set
		blog_id = $2,
		title = $3,
		short_description = $4,
		cover = $5,
		is_free = $6,
		price_rub = $7,
		is_active = $8,
		created = $9,
		updated = $10
		where id = $1
		`
	_, err := r.db.Exec(ctx, query,
		subscription.ID,
		subscription.BlogId,
		subscription.Title,
		subscription.ShortDescription,
		subscription.Cover,
		subscription.IsFree,
		subscription.PriceRub,
		subscription.IsActive,
		subscription.Created,
		subscription.Updated,
	)
	return err
}

func (r *Repository) AllUserSubscriptions(ctx context.Context) ([]UserSubscription, error) {
	query := `select id, user_id, subscription_id, blog_id, status, is_active, expires_at, created, updated 
			from user_subscriptions
			order by blog_id, created desc 
			`

	rows, err := r.db.Query(ctx, query)
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

func (r *Repository) UserSubscriptionsByUserId(ctx context.Context, userId uuid.UUID) ([]UserSubscription, error) {
	query := `select id, user_id, subscription_id, blog_id, status, is_active, expires_at, created, updated 
			from user_subscriptions
			where user_id = $1
			order by blog_id, created desc 
			`

	rows, err := r.db.Query(ctx, query, userId)
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

func (r *Repository) UserSubscriptionsByBlogId(ctx context.Context, blogId uuid.UUID) ([]UserSubscription, error) {
	query := `select id, user_id, subscription_id, blog_id, status, is_active, expires_at, created, updated 
			from user_subscriptions
			where blog_id = $1
			order by created desc 
			`

	rows, err := r.db.Query(ctx, query, blogId)
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

func (r *Repository) UserSubscriptionsByUserIdAndBlogId(ctx context.Context, userId uuid.UUID, blogId uuid.UUID) ([]UserSubscription, error) {
	query := `select id, user_id, subscription_id, blog_id, status, is_active, expires_at, created, updated 
			from user_subscriptions 
			where user_id = $1 and blog_id = $2
			order by created desc`

	rows, err := r.db.Query(ctx, query, userId, blogId)
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

func (r *Repository) UserSubscriptionByParams(ctx context.Context, userId uuid.UUID, blogId uuid.UUID, subscriptionId uuid.UUID) (*UserSubscription, error) {
	query := `select id, user_id, subscription_id, blog_id, status, is_active, expires_at, created, updated 
			from user_subscriptions 
			where user_id = $1 and blog_id = $2 and subscription_id = $3
			order by created desc`

	var subscription UserSubscription
	err := r.db.QueryRow(ctx, query, userId, blogId, subscriptionId).Scan(
		&subscription.ID,
		&subscription.UserId,
		&subscription.SubscriptionId,
		&subscription.BlogId,
		&subscription.Status,
		&subscription.IsActive,
		&subscription.ExpiresAt,
		&subscription.Created,
		&subscription.Updated,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &subscription, nil
}

func (r *Repository) CreateUserSubscription(ctx context.Context, userSubscription *UserSubscription) error {

	query := `insert into user_subscriptions
	(id, user_id, subscription_id, blog_id, status, is_active, expires_at, created, updated)
	values
	($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err := r.db.Exec(ctx, query,
		userSubscription.ID,
		userSubscription.UserId,
		userSubscription.SubscriptionId,
		userSubscription.BlogId,
		userSubscription.Status,
		userSubscription.IsActive,
		userSubscription.ExpiresAt,
		userSubscription.Created,
		userSubscription.Updated,
	)
	return err
}

func (r *Repository) UpdateUserSubscription(ctx context.Context, userSubscription *UserSubscription) error {

	query := `update user_subscriptions set
		user_id = $2,
		subscription_id = $3,
		blog_id = $4,
		status = $5,
		is_active = $6,
		expires_at = $7,
		created = $8,
		updated = $9
		where id = $1
		`
	_, err := r.db.Exec(ctx, query,
		userSubscription.ID,
		userSubscription.UserId,
		userSubscription.SubscriptionId,
		userSubscription.BlogId,
		userSubscription.Status,
		userSubscription.IsActive,
		userSubscription.ExpiresAt,
		userSubscription.Created,
		userSubscription.Updated,
	)
	return err
}

func (r *Repository) CountBlogPaidUserSubscriptions(ctx context.Context, blogId uuid.UUID) (int, error) {

	query := `select count(us.id) from user_subscriptions us
				 join subscriptions s on us.subscription_id = s.id
                 where s.blog_id = $1 and us.is_active = true and s.is_free = false`

	var count int
	err := r.db.QueryRow(ctx, query, blogId).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *Repository) UserFollowsByUserId(ctx context.Context, userId uuid.UUID) ([]UserFollow, error) {

	query := `select id, user_id, blog_id, created, updated
			from user_follows
			where user_id = $1
			order by created desc
			`
	rows, err := r.db.Query(ctx, query, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	resultArray := make([]UserFollow, 0)
	var userFollow UserFollow
	for rows.Next() {
		err = rows.Scan(
			&userFollow.ID,
			&userFollow.UserId,
			&userFollow.BlogId,
			&userFollow.Created,
			&userFollow.Updated,
		)
		if err != nil {
			return nil, err
		}
		resultArray = append(resultArray, userFollow)
	}
	return resultArray, nil
}

func (r *Repository) UserFollowByUserIdAndBlogId(ctx context.Context, userId uuid.UUID, blogId uuid.UUID) (*UserFollow, error) {

	query := `select id, user_id, blog_id, created, updated
			from user_follows
			where user_id = $1 and blog_id = $2
			order by created desc
			`

	var userFollow UserFollow
	err := r.db.QueryRow(ctx, query, userId, blogId).Scan(
		&userFollow.ID,
		&userFollow.UserId,
		&userFollow.BlogId,
		&userFollow.Created,
		&userFollow.Updated,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &userFollow, nil
}

func (r *Repository) CountBlogUserFollows(ctx context.Context, blogId uuid.UUID) (int, error) {
	query := `select count(id) from user_follows where blog_id = $1`
	count := 0
	err := r.db.QueryRow(ctx, query, blogId).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *Repository) CreateUserFollow(ctx context.Context, userFollow *UserFollow) error {

	query := `insert into user_follows
	(id, user_id, blog_id, created, updated)
	values
	($1, $2, $3, $4, $5)`
	_, err := r.db.Exec(ctx, query,
		userFollow.ID,
		userFollow.UserId,
		userFollow.BlogId,
		userFollow.Created,
		userFollow.Updated,
	)
	return err
}

func (r *Repository) CountPostLikes(ctx context.Context, postId uuid.UUID) (int, int, error) {
	query := `select 
    	coalesce(sum(CASE WHEN positive = TRUE THEN 1 ELSE 0 END), 0), 
    	coalesce(sum(CASE WHEN positive = false THEN 1 ELSE 0 END), 0)
	from post_likes where post_id = $1`
	var likes int
	var dislikes int
	err := r.db.QueryRow(ctx, query, postId).Scan(&likes, &dislikes)
	if err != nil {
		return 0, 0, err
	}

	return likes, dislikes, nil
}

func (r *Repository) PostLikeByPostIdAndUserId(ctx context.Context, postId uuid.UUID, userId uuid.UUID) (*PostLike, error) {
	query := `select id, post_id, user_id, positive, created
			from post_likes
			where post_id = $1 and user_id = $2
			`
	var postLike PostLike
	err := r.db.QueryRow(ctx, query, postId, userId).Scan(
		&postLike.ID,
		&postLike.PostId,
		&postLike.UserId,
		&postLike.Positive,
		&postLike.Created,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &postLike, nil
}

func (r *Repository) CreatePostLike(ctx context.Context, postLike *PostLike) error {
	query := `insert into post_likes
	(id, post_id, user_id, positive, created)
	values
	($1, $2, $3, $4, $5)`
	_, err := r.db.Exec(ctx, query,
		postLike.ID,
		postLike.PostId,
		postLike.UserId,
		postLike.Positive,
		postLike.Created,
	)
	return err
}

func (r *Repository) UpdatePostLike(ctx context.Context, postLike *PostLike) error {
	query := `update post_likes
		set post_id = $2, user_id = $3, positive = $4, created = $5
		where id = $1`
	_, err := r.db.Exec(ctx, query,
		postLike.ID,
		postLike.PostId,
		postLike.UserId,
		postLike.Positive,
		postLike.Created,
	)
	return err
}

func (r *Repository) DeletePostLike(ctx context.Context, postLike *PostLike) error {
	query := `delete from post_likes where id = $1`
	_, err := r.db.Exec(ctx, query, postLike.ID)
	return err
}

func (r *Repository) PostPaidAccessByPostIdAndUserId(ctx context.Context, postId uuid.UUID, userId uuid.UUID) (*PostPaidAccess, error) {

	query := `select id, post_id, user_id, created
			from post_paid_access
			where post_id = $1 and user_id = $2
			`
	var postPaidAccess PostPaidAccess
	err := r.db.QueryRow(ctx, query, postId, userId).Scan(
		&postPaidAccess.ID,
		&postPaidAccess.PostId,
		&postPaidAccess.UserId,
		&postPaidAccess.Created,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &postPaidAccess, nil
}

func (r *Repository) CreatePostPaidAccess(ctx context.Context, postPaidAccess *PostPaidAccess) error {

	query := `insert into post_paid_access
	(id, post_id, user_id, created)
	values
	($1, $2, $3, $4)`
	_, err := r.db.Exec(ctx, query,
		postPaidAccess.ID,
		postPaidAccess.PostId,
		postPaidAccess.UserId,
		postPaidAccess.Created,
	)
	return err
}

func (r *Repository) AllGoals(ctx context.Context) ([]Goal, error) {

	query := `select id, blog_id, type, description, target, current, created, updated from goals`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var goals = make([]Goal, 0)
	for rows.Next() {
		var goal Goal
		err := rows.Scan(
			&goal.ID,
			&goal.BlogId,
			&goal.Type,
			&goal.Description,
			&goal.Target,
			&goal.Current,
			&goal.Created,
			&goal.Updated,
		)
		if err != nil {
			return nil, err
		}
		goals = append(goals, goal)
	}

	return goals, nil
}

func (r *Repository) GoalsById(ctx context.Context, id uuid.UUID) (*Goal, error) {
	query := `select id, blog_id, type, description, target, current, created, updated from goals where id = $1`
	var goal Goal
	err := r.db.QueryRow(ctx, query, id).Scan(
		&goal.ID,
		&goal.BlogId,
		&goal.Type,
		&goal.Description,
		&goal.Target,
		&goal.Current,
		&goal.Created,
		&goal.Updated,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &goal, nil

}

func (r *Repository) GoalsByBlogId(ctx context.Context, blogId uuid.UUID) ([]Goal, error) {

	query := `select id, blog_id, type, description, target, current, created, updated from goals where blog_id = $1`
	rows, err := r.db.Query(ctx, query, blogId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var goals = make([]Goal, 0)
	for rows.Next() {
		var goal Goal
		err := rows.Scan(
			&goal.ID,
			&goal.BlogId,
			&goal.Type,
			&goal.Description,
			&goal.Target,
			&goal.Current,
			&goal.Created,
			&goal.Updated,
		)
		if err != nil {
			return nil, err
		}
		goals = append(goals, goal)
	}

	return goals, nil
}

func (r *Repository) CreateGoal(ctx context.Context, goal *Goal) error {

	query := `insert into goals
	(id, blog_id, type, description, target, current, created, updated)
	values
	($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := r.db.Exec(ctx, query,
		goal.ID,
		goal.BlogId,
		goal.Type,
		goal.Description,
		goal.Target,
		goal.Current,
		goal.Created,
		goal.Updated,
	)
	return err
}

func (r *Repository) UpdateGoal(ctx context.Context, goal *Goal) error {

	query := `update goals set
		blog_id = $2,
		type = $3,
		description = $4,
		target = $5,
		current = $6,
		created = $7,
		updated = $8
		where id = $1`
	_, err := r.db.Exec(ctx, query,
		goal.ID,
		goal.BlogId,
		goal.Type,
		goal.Description,
		goal.Target,
		goal.Current,
		goal.Created,
		goal.Updated,
	)
	return err
}

func (r *Repository) PostViewByPostIdAndUserId(ctx context.Context, postId uuid.UUID, userId uuid.UUID) (*PostView, error) {
	query := `select id, post_id, user_id, fingerprint, created 
			from post_views 
			where post_id = $1 and user_id = $2`

	var postView PostView
	err := r.db.QueryRow(ctx, query, postId, userId).Scan(
		&postView.ID,
		&postView.PostId,
		&postView.UserId,
		&postView.Fingerprint,
		&postView.Created,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &postView, nil
}

func (r *Repository) PostViewByPostIdAndFingerprint(ctx context.Context, postId uuid.UUID, fingerprint string) (*PostView, error) {
	query := `select id, post_id, user_id, fingerprint, created 
			from post_views 
			where post_id = $1 and fingerprint = $2`

	var postView PostView
	err := r.db.QueryRow(ctx, query, postId, fingerprint).Scan(
		&postView.ID,
		&postView.PostId,
		&postView.UserId,
		&postView.Fingerprint,
		&postView.Created,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &postView, nil
}

func (r *Repository) CreatePostView(ctx context.Context, postView *PostView) error {

	query := `insert into post_views
	(id, post_id, user_id, fingerprint, created)
	values
	($1, $2, $3, $4, $5)`
	_, err := r.db.Exec(ctx, query,
		postView.ID,
		postView.PostId,
		postView.UserId,
		postView.Fingerprint,
		postView.Created,
	)
	return err
}

func (r *Repository) DonationById(ctx context.Context, id uuid.UUID) (*Donation, error) {
	query := `select id, user_id, blog_id, value, currency, user_comment, status, payment_confirmed, created, updated 
			from donations 
         	where id = $1
         	order by created desc`

	var item Donation
	err := r.db.QueryRow(ctx, query, id).Scan(
		&item.ID,
		&item.UserId,
		&item.BlogId,
		&item.Value,
		&item.Currency,
		&item.UserComment,
		&item.Status,
		&item.PaymentConfirmed,
		&item.Created,
		&item.Updated,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func (r *Repository) DonationsByBlogId(ctx context.Context, blogId uuid.UUID) ([]Donation, error) {
	query := `select id, user_id, blog_id, value, currency, user_comment, status, payment_confirmed, created, updated 
			from donations 
         	where blog_id = $1
         	order by created desc`

	rows, err := r.db.Query(ctx, query, blogId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var donations = make([]Donation, 0)
	var item Donation
	for rows.Next() {
		err := rows.Scan(
			&item.ID,
			&item.UserId,
			&item.BlogId,
			&item.Value,
			&item.Currency,
			&item.UserComment,
			&item.Status,
			&item.PaymentConfirmed,
			&item.Created,
			&item.Updated,
		)
		if err != nil {
			return nil, err
		}
		donations = append(donations, item)
	}
	return donations, nil
}

func (r *Repository) DonationsByBlogIdConfirmed(ctx context.Context, blogId uuid.UUID) ([]Donation, error) {
	query := `select id, user_id, blog_id, value, currency, user_comment, status, payment_confirmed, created, updated 
			from donations 
         	where blog_id = $1 and payment_confirmed = true
         	order by created desc`

	rows, err := r.db.Query(ctx, query, blogId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var donations = make([]Donation, 0)
	var item Donation
	for rows.Next() {
		err := rows.Scan(
			&item.ID,
			&item.UserId,
			&item.BlogId,
			&item.Value,
			&item.Currency,
			&item.UserComment,
			&item.Status,
			&item.PaymentConfirmed,
			&item.Created,
			&item.Updated,
		)
		if err != nil {
			return nil, err
		}
		donations = append(donations, item)
	}
	return donations, nil
}

func (r *Repository) CreateDonation(ctx context.Context, donation *Donation) error {

	query := `insert into donations
	(id, user_id, blog_id, value, currency, user_comment, status, payment_confirmed, created, updated)
	values
	($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
	_, err := r.db.Exec(ctx, query,
		donation.ID,
		donation.UserId,
		donation.BlogId,
		donation.Value,
		donation.Currency,
		donation.UserComment,
		donation.Status,
		donation.PaymentConfirmed,
		donation.Created,
		donation.Updated,
	)
	return err
}

func (r *Repository) UpdateDonation(ctx context.Context, donation *Donation) error {
	query := `update donations set
		user_id = $2,
		blog_id = $3,
		value = $4,
		currency = $5,
		user_comment = $6,
		status = $7,
		payment_confirmed = $8,
		created = $9,
		updated = $10
		where id = $1`
	_, err := r.db.Exec(ctx, query,
		&donation.ID,
		&donation.UserId,
		&donation.BlogId,
		&donation.Value,
		&donation.Currency,
		&donation.UserComment,
		&donation.Status,
		&donation.PaymentConfirmed,
		&donation.Created,
		&donation.Updated,
	)
	return err
}

func (r *Repository) BlogIncomesByBlogId(ctx context.Context, blogId uuid.UUID) ([]BlogIncome, error) {
	query := `select id, blog_id, user_id, value, currency, item_id, item_type, sent_to_user_wallet, created 
			from blog_incomes
		 	where blog_id = $1
		 	order by created desc`

	rows, err := r.db.Query(ctx, query, blogId)
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

func (r *Repository) CreateBlogIncome(ctx context.Context, blogIncome *BlogIncome) error {
	query := `insert into blog_incomes
	(id, blog_id, user_id, value, currency, item_id, item_type, sent_to_user_wallet, created)
	values
	($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err := r.db.Exec(ctx, query,
		blogIncome.ID,
		blogIncome.BlogId,
		blogIncome.UserId,
		blogIncome.Value,
		blogIncome.Currency,
		blogIncome.ItemId,
		blogIncome.ItemType,
		blogIncome.SentToUserWallet,
		blogIncome.Created,
	)
	return err
}

func (r *Repository) UpdateBlogIncome(ctx context.Context, blogIncome *BlogIncome) error {
	query := `update blog_incomes set
		blog_id = $2,
		user_id = $3,
		value = $4,
		currency = $5,
		item_id = $6,
		item_type = $7,
		sent_to_user_wallet = $8,
		created = $9
		where id = $1`
	_, err := r.db.Exec(ctx, query,
		blogIncome.ID,
		blogIncome.BlogId,
		blogIncome.UserId,
		blogIncome.Value,
		blogIncome.Currency,
		blogIncome.ItemId,
		blogIncome.ItemType,
		blogIncome.SentToUserWallet,
		blogIncome.Created,
	)
	return err
}
