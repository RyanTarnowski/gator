package main

import (
	"database/sql"
	"gator/internal/config"
	"gator/internal/database"
	_ "github.com/lib/pq"
	"log"
	"os"
)

type state struct {
	config *config.Config
	db     *database.Queries
}

func main() {
	cfg, err := config.Read()
	if err != nil {
		log.Fatalf("error reading config: %v", err)
	}

	db, err := sql.Open("postgres", cfg.DbURL)
	if err != nil {
		log.Fatalf("error opening database: %v", err)
	}
	defer db.Close()
	dbQueries := database.New(db)

	programState := &state{
		config: &cfg,
		db:     dbQueries,
	}

	cmds := commands{
		commands: make(map[string]func(*state, command) error),
	}
	cmds.register("login", handlerLogin)
	cmds.register("register", handlerRegister)
	cmds.register("reset", handlerReset)
	cmds.register("users", handlerGetUsers)
	cmds.register("agg", handlerAgg)
	cmds.register("addfeed", middlewareLoggedIn(handlerAddFeed))
	cmds.register("feeds", handlerFeeds)
	cmds.register("follow", middlewareLoggedIn(handlerFollow))
	cmds.register("following", middlewareLoggedIn(handlerFollowing))
	cmds.register("unfollow", middlewareLoggedIn(handlerUnfollow))

	if len(os.Args) < 2 {
		log.Fatal("Usage: cli <command> [args...]")
	}

	cmdName := os.Args[1]
	cmdArgs := os.Args[2:]

	err = cmds.run(programState, command{name: cmdName, args: cmdArgs})
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
