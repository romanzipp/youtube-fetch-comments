package main

import (
	"bufio"
	"context"
	"encoding/csv"
	"fmt"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
	"google.golang.org/api/youtube/v3"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"google.golang.org/api/option"
)

type Comment struct {
	Id         string
	AuthorName string
	Text       string
	IsReply    bool
	Replies    []Comment
}

type VideoInfo struct {
	VideoId        string
	Comments       []Comment
	CommentThreads []*youtube.CommentThread
}

func getEnvOrDefaultInt(key string, fallback int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return fallback
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		log.Fatalf("failed to convert %s to integer: %s", key, err)
	}

	return value
}

func padRight(str string, length int) string {
	return fmt.Sprintf("%-*s", length, str)
}

func main() {
	ctx := context.Background()

	minLength := getEnvOrDefaultInt("MIN_LENGTH", 5)
	maxComments := getEnvOrDefaultInt("MAX_COMMENTS", 1000)

	apiKey := os.Getenv("YOUTUBE_API_KEY")

	if apiKey == "" {
		log.Fatalf("YOUTUBE_API_KEY is required")
	}

	fmt.Println("[App] api key      =", apiKey)
	fmt.Println("[App] min length   =", minLength)
	fmt.Println("[App] max comments =", maxComments)

	file, err := os.Open("videos.txt")
	if err != nil {
		log.Fatalf("failed to open file: %s", err)
	}

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	var videoIds []string

	for scanner.Scan() {
		videoUrl := scanner.Text()
		videoId, err := getYoutubeID(videoUrl)
		if err != nil {
			log.Fatalf("failed to find id in string: %s", videoUrl)
		}
		videoIds = append(videoIds, videoId)
	}

	file.Close()

	service, err := youtube.NewService(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Fatalf("Error creating new YouTube client: %v", err)
	}

	videoInfos := make([]VideoInfo, 0)

	for _, videoId := range videoIds {
		fmt.Println("[Comments] fetching for video", videoId)

		info := VideoInfo{VideoId: videoId}

		call := service.CommentThreads.List([]string{"snippet"}).VideoId(videoId).MaxResults(100)
		count := 0

		for {

			err = call.Pages(ctx, func(response *youtube.CommentThreadListResponse) error {
				fmt.Println("[Comments] fetching comments", count)

				for _, item := range response.Items {
					// skip short comments
					if len(item.Snippet.TopLevelComment.Snippet.TextDisplay) < minLength {
						continue
					}

					// remember comments which have replies
					if item.Snippet.TotalReplyCount > 0 {
						info.CommentThreads = append(info.CommentThreads, item)
					}

					info.Comments = append(info.Comments, Comment{
						Id:         item.Id,
						AuthorName: item.Snippet.TopLevelComment.Snippet.AuthorDisplayName,
						Text:       item.Snippet.TopLevelComment.Snippet.TextDisplay,
					})
				}

				count += len(response.Items)
				if count >= maxComments {
					fmt.Println("[Comments] fetching comments - reached max items")
					return fmt.Errorf("reached max items")
				}

				return nil
			})

			if err != nil {
				if strings.Contains(err.Error(), "too much requests") || strings.Contains(err.Error(), "quotaExceeded") || strings.Contains(err.Error(), "rateLimitExceeded") {
					fmt.Println("[Comments] WARNING: too much requests, waiting for 10 seconds before retrying...")
					time.Sleep(10 * time.Second)
					continue
				} else if strings.Contains(err.Error(), "commentsDisabled") {
					fmt.Println("[Comments] WARNING: comments are disabled for this video")
				} else if err.Error() != "reached max items" {
					log.Fatalf("Error fetching comments: %v", err)
				}
			}

			break
		}

		videoInfos = append(videoInfos, info)
	}

	for _, info := range videoInfos {
		for _, thread := range info.CommentThreads {
			call := service.Comments.List([]string{"snippet"}).ParentId(thread.Snippet.TopLevelComment.Id).MaxResults(100)

			for {
				err = call.Pages(ctx, func(response *youtube.CommentListResponse) error {
					fmt.Println("[Comment-Replies] fetching reply comments for video", info.VideoId, "+ comment", thread.Snippet.TopLevelComment.Id)

					// got replies for comment thread
					for _, item := range response.Items {
						// skip short comments
						if len(item.Snippet.TextDisplay) < minLength {
							continue
						}

						reply := Comment{
							AuthorName: item.Snippet.AuthorDisplayName,
							Text:       item.Snippet.TextDisplay,
							IsReply:    true,
						}

						// Find the parent comment and append the reply to its Replies field
						for i := range info.Comments {
							if info.Comments[i].Id == item.Snippet.ParentId {
								info.Comments[i].Replies = append(info.Comments[i].Replies, reply)
								break
							}
						}
					}

					return nil
				})

				if err != nil {
					if strings.Contains(err.Error(), "too much requests") || strings.Contains(err.Error(), "quotaExceeded") || strings.Contains(err.Error(), "rateLimitExceeded") {
						fmt.Println("[Comment-Replies] WARNING: too much requests, waiting for 10 seconds before retrying...")
						time.Sleep(10 * time.Second)
						continue
					}

					log.Fatalf("Error fetching reply comments: %v", err)
				}

				break
			}

			info.CommentThreads = nil
		}
	}

	records := [][]string{
		{"videoId", "isReply", "authorName", "commentText"},
	}

	for _, info := range videoInfos {
		for _, comment := range info.Comments {
			root := "."
			if len(comment.Replies) > 0 {
				root = "┌───"
			}

			records = append(records, []string{info.VideoId, root, cleanCol(comment.AuthorName), cleanCol(comment.Text)})

			// lines = append(lines, fmt.Sprintf("%s;%s;%s;%s", info.VideoId, root, padRight(cleanCol(comment.AuthorName), 30), cleanCol(comment.Text)))

			for replyIndex, replyComment := range comment.Replies {
				angle := "├"
				if replyIndex == len(comment.Replies)-1 {
					angle = "└"
				} else {
					angle = "├"
				}

				records = append(records, []string{info.VideoId, fmt.Sprintf("%s%s", angle, "── "), cleanCol(replyComment.AuthorName), cleanCol(replyComment.Text)})
				//lines = append(lines, fmt.Sprintf("%s;%s%s;%s;%s", info.VideoId, angle, "── ", cleanCol(replyComment.AuthorName), cleanCol(replyComment.Text)))
			}
		}
	}

	f, _ := os.Create("comments.csv")
	t := transform.NewWriter(f, unicode.UTF8BOM.NewEncoder())
	w := csv.NewWriter(t)
	w.Comma = ';'

	for _, record := range records {
		w.Write(record)
	}

	w.Flush()
}

func cleanCol(text string) string {
	return strings.ReplaceAll(strings.ReplaceAll(text, ";", ","), "\n", " ")
}

func getYoutubeID(urlStr string) (string, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return "", err
	}

	if u.Host == "youtu.be" {
		return strings.TrimPrefix(u.Path, "/"), nil
	}

	if u.Host == "www.youtube.com" {
		q, err := url.ParseQuery(u.RawQuery)
		if err != nil {
			return "", err
		}

		return q.Get("v"), nil
	}

	return "", fmt.Errorf("unknown host: %s", u.Host)
}
