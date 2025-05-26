package main

type ProcessedPosts struct {
	Posts []Post
	Tags  map[string][]int
}

func processPosts(posts []Post) ProcessedPosts {
	processedPosts := ProcessedPosts{}
	processedPosts.Tags = make(map[string][]int)

	for _, post := range posts {
		processedPosts.Posts = append(processedPosts.Posts, post)

		for _, tag := range post.FrontMatter.Tags {
			processedPosts.Tags[tag] = append(processedPosts.Tags[tag], post.Index)
		}
	}

	return processedPosts
}
