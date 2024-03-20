# Fetch YouTube Comments

## Usage

### 1. Create a `videos.txt`

Create a `videos.txt` file with the YouTube video URLs separated by new line you want to fetch the comments from.

```txt
https://youtu.be/h5MJn_Yy7aA
https://www.youtube.com/watch?v=eglxpDucXpI
...
```

### 2. Call the script

You can/need to pass the following env vars:

- `YOUTUBE_API_KEY`: Your YouTube API key
- `MIN_LENGTH`: Minimum length of the comment to be fetched (default: 5)
- `MAX_COMMENTS`: Maximum amount of the comment to be fetched for each video (default: 1000)


```
YOUTUBE_API_KEY=... MIN_LENGTH=10 ./fetch_comments
```

### 3. Output file

If everything goes to plan, you should see a `comments.csv` file being created.

## Build

### Development

```shell
go run .
```

### Production

```shell
go mod download
go build -o ./fetch_comments
```
