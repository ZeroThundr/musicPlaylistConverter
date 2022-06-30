package main

import (
	"fmt"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/youtube/v3"
	"io/ioutil"
	"log"
	"math"
	"net/http"
)

func getGoogleClient(scope string) *http.Client {
	ctx := context.Background()

	b, err := ioutil.ReadFile("googleClientSecret.json")
	if err != nil {
		log.Fatalf("Unable to read google client secret file: %v", err)
	}

	// If modifying the scope, delete your previously saved credentials
	// at ~/.credentials/youtube-go.json
	config, err := google.ConfigFromJSON(b, scope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}

	// Use a redirect URI like this for a web app. The redirect URI must be a
	// valid one for your OAuth2 credentials.
	config.RedirectURL = "http://localhost:8080"
	// Use the following redirect URI if launchWebServer=false in oauth2.go
	// config.RedirectURL = "urn:ietf:wg:oauth:2.0:oob"

	cacheFile, err := tokenCacheFile("youtube")
	if err != nil {
		log.Fatalf("Unable to get path to cached credential file. %v", err)
	}
	tok, err := tokenFromFile(cacheFile)
	if err != nil {
		authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
		if launchWebServer {
			fmt.Println("Trying to get token from web")
			tok, err = getTokenFromWeb(config, authURL)
		} else {
			fmt.Println("Trying to get token from prompt")
			tok, err = getTokenFromPrompt(config, authURL)
		}
		if err == nil {
			saveToken(cacheFile, tok)
		}
	}
	return config.Client(ctx, tok)
}
func getYoutubeVideoID(service *youtube.Service, videoName string) *youtube.SearchListResponse { //gets individual video IDs
	part := []string{"id,snippet"}
	call := service.Search.List(part).
		Q(videoName).
		MaxResults(1)
	response, err := call.Do()
	handleError(err, "")
	return response
}
func youtubePlaylistMaker(service *youtube.Service, part []string, playlistName *youtube.PlaylistSnippet) string { //creates an empty playlist and returns the ID
	playlist := &youtube.Playlist{
		Snippet: playlistName,
	}
	call := service.Playlists.Insert(part, playlist)
	response, err := call.Do()
	handleError(err, "")
	return response.Id
}
func addItemsToYoutubePlaylist(service *youtube.Service, playListId string, videoId string) *youtube.PlaylistItem { //adds videos to playlist
	part := []string{"id,snippet"}
	resourceId := &youtube.ResourceId{
		Kind:    "youtube#video",
		VideoId: videoId,
	}
	videoResourceSnippet := &youtube.PlaylistItemSnippet{
		PlaylistId: playListId,
		ResourceId: resourceId,
	}
	videoResource := &youtube.PlaylistItem{
		Snippet: videoResourceSnippet,
	}
	call := service.PlaylistItems.Insert(part, videoResource)
	response, err := call.Do()
	handleError(err, "")
	return response
}
func playlistItemsList(service *youtube.Service, part []string, playlistId string, pageToken string) *youtube.PlaylistItemListResponse { //grabs all the items in a YouTube playlist
	call := service.PlaylistItems.List(part)
	call = call.PlaylistId(playlistId)
	if pageToken != "" {
		call = call.PageToken(pageToken)
	}
	response, err := call.Do()
	handleError(err, "")
	return response
}

func createYouTubePlaylist(playlist []string) { //creates a YouTube playlist and adds the songs listed in the playlist slice to it
	var part = []string{"id,snippet"}
	var videoIdList []string
	client := getGoogleClient(youtube.YoutubepartnerScope)
	service, err := youtube.New(client)
	if err != nil {
		log.Fatalf("Error creating YouTube client: %v", err)
	}
	for i := range playlist { //gets Video IDs of songs by using the YouTube search method
		videoSearch := getYoutubeVideoID(service, playlist[i])
		for _, item := range videoSearch.Items {
			videoIdList = append(videoIdList, item.Id.VideoId)
		}
	}
	if len(videoIdList) < 200 { //YouTube has a 200 video per playlist limit, this splits the songs into multiple playlists if it is bigger then 200
		playlistDetails := &youtube.PlaylistSnippet{
			Title: "Converted Playlist",
		}
		ytPlaylistId := youtubePlaylistMaker(service, part, playlistDetails)
		for x := range videoIdList {
			addItemsToYoutubePlaylist(service, ytPlaylistId, videoIdList[x])
		}
	} else {
		i := 0
		x := 0
		name := fmt.Sprintf("%v%d", "Playlist #", i+1)
		howManyTimes := int(math.Ceil(float64(len(videoIdList) / 200)))
		lenVideoIdList := len(videoIdList)
		for i < howManyTimes {
			c := 0
			playlistDetails := &youtube.PlaylistSnippet{
				Title: name,
			}
			i += 1
			ytPlaylistId := youtubePlaylistMaker(service, part, playlistDetails)
			for x < lenVideoIdList {
				if c < 200 {
					addItemsToYoutubePlaylist(service, ytPlaylistId, videoIdList[x])
					c += 1
					x += 1
				} else {
					break
				}
			}
		}
	}
	fmt.Println(videoIdList[0])
	fmt.Println("created youtube playlist")
	return
}
