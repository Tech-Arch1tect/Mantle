package main

type PostProcessor interface {
	Process(posts []Post) ProcessedPosts
}

type ProcessedPosts struct {
	Posts []Post           `json:"posts"`
	Tags  map[string][]int `json:"tags"`
}

type DefaultPostProcessor struct{}

func NewPostProcessor() PostProcessor {
	return &DefaultPostProcessor{}
}

func (pp *DefaultPostProcessor) Process(posts []Post) ProcessedPosts {
	processedPosts := ProcessedPosts{
		Posts: make([]Post, 0, len(posts)),
		Tags:  make(map[string][]int),
	}

	for _, post := range posts {
		processedPosts.Posts = append(processedPosts.Posts, post)

		for _, tag := range post.FrontMatter.Tags {
			processedPosts.Tags[tag] = append(processedPosts.Tags[tag], post.Index)
		}
	}

	return processedPosts
}
