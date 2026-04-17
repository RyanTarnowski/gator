package main

import (
	"context"
	"errors"
	"fmt"
	"gator/internal/database"
	"github.com/google/uuid"
	"log"
	"strconv"
	"strings"
	"time"
)

type command struct {
	name string
	args []string
}

type commands struct {
	commands map[string]func(*state, command) error
}

func (c *commands) run(s *state, cmd command) error {
	run_cmd, ok := c.commands[cmd.name]
	if !ok {
		return errors.New("command not found")
	}

	return run_cmd(s, cmd)
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.commands[name] = f
}

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
	return func(s *state, c command) error {
		user, err := s.db.GetUser(context.Background(), s.config.CurrentUserName)
		if err != nil {
			return fmt.Errorf("Was not able to get user: %w", err)
		}
		return handler(s, c, user)
	}
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("Login requires a username input")
	}

	username := cmd.args[0]

	user, err := s.db.GetUser(context.Background(), username)
	if err != nil {
		return fmt.Errorf("Was not able to get user: %w", err)
	}

	err = s.config.SetUser(user.Name)
	if err != nil {
		return fmt.Errorf("Not able to set current user: %w", err)
	}

	fmt.Printf("User %v has been logged in.\n", username)
	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("Register requires a username input")
	}

	createUserParams := database.CreateUserParams{
		ID:        uuid.New(),
		Name:      cmd.args[0],
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	user, err := s.db.CreateUser(context.Background(), createUserParams)
	if err != nil {
		return fmt.Errorf("Failed to register user: %w", err)
	}

	err = s.config.SetUser(user.Name)
	if err != nil {
		return fmt.Errorf("Not able to set current user: %w", err)
	}

	fmt.Printf("User %v has been register and is now logged in.\n", user.Name)
	return nil
}

func handlerReset(s *state, _ command) error {
	err := s.db.ClearUsers(context.Background())
	if err != nil {
		return fmt.Errorf("Failed to clear users: %w", err)
	}

	fmt.Println("User table cleared.")
	return nil
}

func handlerGetUsers(s *state, _ command) error {
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("Failed to get users: %w", err)
	}

	for _, user := range users {
		if user.Name == s.config.CurrentUserName {
			fmt.Printf("* %s (current)\n", user.Name)
		} else {
			fmt.Printf("* %s\n", user.Name)
		}
	}
	return nil
}

func handlerAgg(s *state, cmd command) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("Agg requires a time between reqs value")
	}

	timeBetweenRequests, err := time.ParseDuration(cmd.args[0])
	if err != nil {
		return fmt.Errorf("Failed to parse duration: %w", err)
	}

	fmt.Printf("Collecting feeds every %s...\n", timeBetweenRequests)
	ticker := time.NewTicker(timeBetweenRequests)
	for ; ; <-ticker.C {
		scrapeFeeds(s, cmd)
	}
}

func scrapeFeeds(s *state, _ command) {
	nextFeed, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		log.Printf("Failed to get next feed: %v", err)
		return
	}

	err = s.db.MarkFeedFetched(context.Background(), nextFeed.ID)
	if err != nil {
		log.Printf("Failed to mark feed as fetched: %v", err)
		return
	}

	feed, err := fetchFeed(context.Background(), nextFeed.Url)
	if err != nil {
		log.Printf("Failed to fetch feed: %v", err)
		return
	}

	fmt.Printf("Feed: %v Post Count: %v\n", nextFeed.Name, len(feed.Channel.Item))
	fmt.Println("----------------------------------------------------------------------------------------------------------")
	for _, item := range feed.Channel.Item {

		publishedAt, err := time.Parse(time.RFC1123, item.PubDate)
		if err != nil {
			log.Printf("Error converting publishedAt: %v", err)
			return
		}

		createPostParams := database.CreatePostParams{
			ID:          uuid.New(),
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
			Title:       item.Title,
			Url:         item.Link,
			Description: item.Description,
			PublishedAt: publishedAt,
			FeedID:      nextFeed.ID,
		}

		_, err = s.db.CreatePost(context.Background(), createPostParams)
		if err != nil {
			if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
				continue
			}
			log.Printf("Error saving post: %v", err)
			continue
		}
	}
}

func handlerAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.args) != 2 {
		return fmt.Errorf("Add feed requires a name and url")
	}

	createFeedParams := database.CreateFeedParams{
		ID:        uuid.New(),
		Name:      cmd.args[0],
		Url:       cmd.args[1],
		UserID:    user.ID,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	feed, err := s.db.CreateFeed(context.Background(), createFeedParams)
	if err != nil {
		return fmt.Errorf("Failed to add feed: %w", err)
	}

	createFeedFollowParams := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		UserID:    user.ID,
		FeedID:    feed.ID,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	feed_follow, err := s.db.CreateFeedFollow(context.Background(), createFeedFollowParams)
	if err != nil {
		return fmt.Errorf("Failed to follow feed: %w", err)
	}

	fmt.Println("Feed Added:")
	fmt.Printf("* ID: %v\n", feed.ID)
	fmt.Printf("* Name: %v\n", feed.Name)
	fmt.Printf("* Url: %v\n", feed.Url)
	fmt.Printf("* UserID: %v\n", feed.UserID)
	fmt.Printf("* CreatedAt: %v\n", feed.CreatedAt)
	fmt.Printf("* UpdatedAt: %v\n", feed.UpdatedAt)

	fmt.Println()

	fmt.Println("Feed Followed:")
	fmt.Println("--------------------------------------------------------------------")
	fmt.Printf("Feed Name: %s\n", feed_follow.FeedName)
	fmt.Printf("User Name: %s\n", feed_follow.UserName)

	return nil
}

func handlerFeeds(s *state, _ command) error {
	feeds, err := s.db.GetFeeds(context.Background())
	if err != nil {
		return fmt.Errorf("Was not able to get feeds: %w", err)
	}

	if len(feeds) == 0 {
		fmt.Println("No feeds found.")
		return nil
	}

	fmt.Printf("Found %d feeds:\n", len(feeds))
	for _, feed := range feeds {
		fmt.Println("--------------------------------------------------------------------")
		fmt.Printf("Feed Name: %s\n", feed.FeedName)
		fmt.Printf("URL: %s\n", feed.Url)
		fmt.Printf("User Name: %s\n", feed.UserName)
	}

	return nil
}

func handlerFollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("Following a feed requires a url")
	}

	feed, err := s.db.GetFeed(context.Background(), cmd.args[0])
	if err != nil {
		return fmt.Errorf("Was not able to get url: %w", err)
	}

	createFeedFollowParams := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		UserID:    user.ID,
		FeedID:    feed.ID,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	feed_follow, err := s.db.CreateFeedFollow(context.Background(), createFeedFollowParams)
	if err != nil {
		return fmt.Errorf("Failed to follow feed: %w", err)
	}

	fmt.Println("Following Feed:")
	fmt.Printf("* Feed Name: %v\n", feed_follow.FeedName)
	fmt.Printf("* User Name: %v\n", feed_follow.UserName)
	return nil
}

func handlerFollowing(s *state, cmd command, user database.User) error {
	feed_follows, err := s.db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		return fmt.Errorf("Failed get followed feeds: %w", err)
	}

	if len(feed_follows) == 0 {
		fmt.Println("No feed follows found for this user.")
		return nil
	}

	fmt.Printf("Following %d feeds:\n", len(feed_follows))
	for _, feed := range feed_follows {
		fmt.Println("--------------------------------------------------------------------")
		fmt.Printf("Feed Name: %s\n", feed.FeedName)
		fmt.Printf("User Name: %s\n", feed.UserName)
	}

	return nil
}

func handlerUnfollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("Folloing a feed requires a url")
	}

	feed, err := s.db.GetFeed(context.Background(), cmd.args[0])
	if err != nil {
		return fmt.Errorf("Was not able to get url: %w", err)
	}

	removeFeedFollowParams := database.RemoveFeedFollowParams{
		UserID: user.ID,
		FeedID: feed.ID,
	}

	err = s.db.RemoveFeedFollow(context.Background(), removeFeedFollowParams)
	if err != nil {
		return fmt.Errorf("Error removing feed follow: %w", err)
	}

	fmt.Println("Feed follow removed.")
	return nil
}

func handlerBrowse(s *state, cmd command, user database.User) error {
	limit := int32(2)

	if len(cmd.args) == 1 {
		i64, err := strconv.ParseInt(cmd.args[0], 10, 32)
		if err != nil {
			return fmt.Errorf("Error converting arg to int: %w", err)
		}

		limit = int32(i64)
	}

	getPostsForUserParams := database.GetPostsForUserParams{
		UserID: user.ID,
		Limit:  limit,
	}

	posts, err := s.db.GetPostsForUser(context.Background(), getPostsForUserParams)
	if err != nil {
		return fmt.Errorf("Error getting posts: %w", err)
	}

	for _, post := range posts {
		fmt.Println("--------------------------------------------------------------------")
		fmt.Printf("Feed: %s\n", post.FeedName)
		fmt.Printf("Title: %s\n", post.Title)
		fmt.Printf("URL: %s\n", post.Url)
		fmt.Printf("Published At: %s\n", post.PublishedAt)
		fmt.Printf("Description: %s\n", post.Description)
		fmt.Println("--------------------------------------------------------------------")
		fmt.Println()
	}

	return nil
}
