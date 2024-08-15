# shieldbot
This bot provides non-standard moderation features, such as an isolation and restore command. 
Isolation and restore both require manage roles. Configuring the bot beyond that requires administrator in the relevant server.
It has been designed to be expandable, but is currently basic and mostly framework.
Maintained and updated by @shininglegend.

## Installation
Clone the repo to local: `git clone https://github.com/shininglegend/shieldbot.git`

Copy and rename the copy of `config sample.yaml` to `config.yaml`

Make a database file: `sqlite3 ./data/main.db` 
*Note: Update the config.yaml if you change the path!*

Put your bot token in `config.yaml`

Ensure go is installed. See https://go.dev/doc/install if needed.

## Usage
```bash
go run cmd/bot/main.go
```

## Contributing:
If you make changes, open a pull request! Be clear on what exactly you changed.
You may also use issues to discuss issues, or dm me on discord if you have problems.

### Notes:
Much of the code is written by Claude.ai. 
I program as a day job, but I'm also a bit lazy ;) 
*And no, I have no fear of losing my job to Ai anytime soon.*
