package models

type (

	// Reddit API models

	Listing struct {
		Kind string
		Data ListingData
	}
	ListingData struct {
		After    *string
		Before   *string
		Children []Link
	}
	Link struct {
		Kind string
		Data LinkData
	}
	LinkData struct {
		Name           string
		Subreddit      string
		AuthorFullname string `json:"author_fullname"`
		Title          string
		Author         string
		Ups            int
	}
	Stats struct {
		Posts []LinkStats
		Users []UserStats
	}

	// Service API models

	LinkStats struct {
		Name    string
		Title   string
		Author  string
		UpVotes int
	}
	UserStats struct {
		Name      string
		PostCount int
	}
)
