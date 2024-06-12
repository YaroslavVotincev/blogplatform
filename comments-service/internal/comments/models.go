package comments

import (
	"github.com/google/uuid"
	"time"
)

type Comment struct {
	ID       uuid.UUID `json:"id"`
	ParentId uuid.UUID `json:"parent_id"`
	AuthorId uuid.UUID `json:"author_id"`
	Children []Comment `json:"children"`
	Content  string    `json:"content"`
	Created  time.Time `json:"created"`
	Updated  time.Time `json:"updated"`
}
