# Gator

This project was part of the Backend Developer learning path on Boot.dev.

The goal of this project was to create a CLI RSS feed aggregator.

## Required Packages
- Go
- Postgres

## Setup config file
1. Create a new file called ".gatorconfig.json" at the root of your home directory
2. The JSON file should be formatted like this:
   ```
   {
     "db_url": "postgres://postgres:postgres@localhost:5432/gator?sslmode=disable",
     "current_user_name": ""
   }
   ```
   - db_url: should contain your postgres database connection string
   - current_user_name: contains the current user logged into Gator, this can remain blank for now

## How to Install
1. From the root of the project, run the following command:
   ```
   go install
   ```
   
## How to Run
1. From the root of the project, run the following command:
   ```
   gator [Command] [Args]
   ```
   
## Gator Commands
- login
  - Takes one argument, the username, and "logs" them into Gator (Writes username to .gatorconfig.json)
- register
  - Takes one argument, the username, and inserts their username into the database. This command will also log in the user
- reset
  - **Danger** clear all data from database tables
- users
  - Lists all users set up in the database
- agg
  - Starts a continuous loop to scrape posts from feeds and save the posts to the database. Press Ctrl+C to stop
- addfeed
  - Takes two arguments, feed name and feed URL, and inserts the feed data into the database  
- feeds
  - Lists all feeds set up in the database
- follow
  - Takes one argument, feed URL (needs to exist in Gator, use addfeed if not set up), and inserts the follow data into the database
- following
  - Lists all the followed feeds for the current user   
- unfollow
- browse
