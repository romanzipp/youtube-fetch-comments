# Fetch YouTube Comments

## Usage

### 1. Download executable

[Download the latest executable](https://github.com/romanzipp/youtube-fetch-comments/releases/latest) for your system.

- Windows: `fetch_comments_windows.exe`
- macOS: `fetch_comments_mac`
- Linux: `fetch_comments_arm64`

If you get a  `permissions denied` error, you may call `sudo chmod +x fetch_comments`.

### 2. Create a `videos.txt`

Create a `videos.txt` file with the YouTube video URLs separated by new line you want to fetch the comments from **in the same folder as the executable**.

```txt
https://youtu.be/h5MJn_Yy7aA
https://www.youtube.com/watch?v=eglxpDucXpI
...
```

### 3. Call the script

You can/need to pass the following env vars:

- `YOUTUBE_API_KEY`: Your YouTube API key
- `MIN_LENGTH`: Minimum length of the comment to be fetched (default: 5)
- `MAX_COMMENTS`: Maximum amount of the comment to be fetched for each video (default: 1000)

#### Windows (cmd)

```
set YOUTUBE_API_KEY=... set MIN_LENGTH=10 fetch_comments.exe
```

#### Windows (PowerShell)

```
$env:YOUTUBE_API_KEY="..." $env:MIN_LENGTH="10" .\fetch_comments.exe
```

#### Unix-based systems (macOS, Linux)

```
YOUTUBE_API_KEY=... MIN_LENGTH=10 ./fetch_comments
```

### 4. Output file

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
