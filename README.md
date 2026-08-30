# Gator

 Gator is a boot.dev project intended to practice using go coding and the postgres sql database.
## What Gator does:
```Gator is short for Aggregator. It aggregates items from RSS Feeds. Users who register with the program, set up RSS feeds by loading the URLs into the database and telling Gator to scrape the feeds for new items. Users can then browse the items in the database.``` 

### To compile gator 
'''you will first need to install the go programming language from the official go site - download and install go. You will then need to install postgreSQL, go to the official postgreSQL download page. You can then compile the go code from your repo by running > go build. This will create a binary file gator'''

### To run gator
Run the gator binary file in a Terminal window by entering > ./gator <command> <arguments>

### gator usage, commands

* register <name>  will register <name> and login <name> setting <name> as the current user.
* login <name> logs in <name> assuming you have registered multiple users and will need to switch between users
* users <> no argument, lists all the users that have been registered and will indicate the current user logged in.
* feeds <> no argument, lists all the feeds that have been added.
* reset <> no argument, resets all users and resets the database tables to their null state.

* the following commands require a user to be logged in.

* addfeed <feed_name> <url>  adds a feed and creates a feed_follow for the current ie user who is logged in.
* follow <url> allows the logged in user to follow a feed that another user has added
* unfollow <url> does exactly what you would expect for the current logged in user
* agg <Ns, or Nm, or Nh> will start aggragating feeds for the current user for all the feeds that user is following every N seconds, or N minutes or N hours. N is an integer value
* browse <M> allows the logged in user to browse through up to M number of feeds that have been aggragated into the database for that user. M is an integer value.

* You will need to give gator the number of arguments it expects for each command otherwise you will get an error message.

### Finding RSS Feeds
'''You can find RSS feeds from simple searches on the internet. As long as the URL is correct you can give a feed any name you like, just enter a feed_name as a single string.'''

### gator tools

The gator project made use of a number of tools to facilitate the process of generating go code.

### sqlc 
'''Is used to generate go code from SQL statements. In this project it greatly simplified the process of creating functions that executed queries on the database.'''

### goose 
'''Is a SQL to database migration management tool. The project used goose to add successive tables to the database. To be able to troubleshoot if changes to the database structure created errors with the code, goose makes use of up migrations and down migrations. For every SQL statement that created a new table, including tables that created relationships between existing tables, code was created for the "up migration" to add the new table and code was created for the "down migration" to remove that new table.'''
