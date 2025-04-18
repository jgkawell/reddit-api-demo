# reddit-api-demo

A Reddit API Demo program that subscribes to your configured subreddits and returns a sorted list of posts with most upvotes and users with most posts to the given subreddit while the program is running.

## How to Run

Before running, you'll need to add your `accessToken` within the `config.yaml` file. You may also wish to configure configure the `reddit.subreddits` array to track whichever subreddits you like. Here are your options:

- The `name` value is required and is just the name of the subreddit you wish to track stats for (e.g. `funny`).
- The `start` value optional and will collect stats on all posts after the specified one. This value is the `fullname` of a Reddit post (e.g. `t3_15bfi0`). If not set, the program will track all new posts that are published after the program starts up so depending on your choice of subreddit(s) you may have to wait a few minutes for data to populate.
- The options under `service` are all good as they are but you may want to change the `logging.level` (options are `error`, `warn`, `info`, `debug`, and `trace`) and and the `network.bind_port`.

Once everything is configured you can run the program with:

```sh
go run main.go
```

The program will bind to `localhost:8080` by default. You can then request data from the built in API using anything you'd like. The server is listening on `/api/stats` and the parameters:

- `sub <string>`: the subreddit to return the stats for
- `limit <int>`: the limit of posts and users to return (optional)

So for example, you could get the data using curl with:

```sh
curl 'localhost:8080/api/stats?sub=funny&limit=15'
```

Or follow the changes with:

```sh
watch curl 'localhost:8080/api/stats?sub=funny&limit=15'
```



