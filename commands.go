package main

import (
	"context"
	"errors"
	"fmt"
	"gator/internal/database"
	"github.com/google/uuid"
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

func handlerAgg(_ *state, _ command) error {
	rss_feed, err := fetchFeed(context.Background(), "https://www.wagslane.dev/index.xml")
	if err != nil {
		return fmt.Errorf("Failed to get RSS Feed: %w", err)
	}

	for _, item := range rss_feed.Channel.Item {
		fmt.Printf("Title: %s\n", item.Title)
		fmt.Printf("Link: %s\n", item.Link)
		fmt.Printf("Desc: %s\n", item.Description)
		fmt.Printf("Date: %s\n", item.PubDate)
	}

	return nil
}

func handlerAddFeed(s *state, cmd command) error {
	if len(cmd.args) != 2 {
		return fmt.Errorf("Add feed requires a name and url")
	}

	user, err := s.db.GetUser(context.Background(), s.config.CurrentUserName)
	if err != nil {
		return fmt.Errorf("Was not able to get user: %w", err)
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

func handlerFollow(s *state, cmd command) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("Folloing a feed requires a url")
	}

	user, err := s.db.GetUser(context.Background(), s.config.CurrentUserName)
	if err != nil {
		return fmt.Errorf("Was not able to get user: %w", err)
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

func handlerFollowing(s *state, cmd command) error {
	user, err := s.db.GetUser(context.Background(), s.config.CurrentUserName)
	if err != nil {
		return fmt.Errorf("Was not able to get user: %w", err)
	}

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
