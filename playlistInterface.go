package main

import (
	"fmt"
	"github.com/zmb3/spotify"
	"google.golang.org/api/youtube/v3"
	"log"
	"strings"
)

type Playlist interface {
	GetSongs() []string
}
type YouTube struct {
}

func NewYoutube() Playlist {
	return &YouTube{}
}
func (Y *YouTube) GetSongs() []string {
	var playlist []string
	part := []string{"snippet"}
	undesiredVideoTitles := []string{"[official music video]", "[official lyric video]", "[official video]", "[official audio]", "[audio]", "[video]", "[animated music video]", "(official music video)", "(official video)", "(official audio)", "(audio)", "(video)", "(animated music video)"}
	client := getGoogleClient(youtube.YoutubeReadonlyScope)
	service, err := youtube.New(client)

	if err != nil {
		log.Fatalf("Error creating YouTube client: %v", err)
	}

	playlistId := playlistIDFromURL("youtube") // Print the playlist ID for the list of uploaded videos.
	fmt.Printf("Videos in list %s\r\n", playlistId)

	nextPageToken := ""
	for {
		// Retrieve next set of items in the playlist.
		playlistResponse := playlistItemsList(service, part, playlistId, nextPageToken)

		for _, playlistItem := range playlistResponse.Items {
			title := strings.ToLower(playlistItem.Snippet.Title)
			videoId := playlistItem.Snippet.ResourceId.VideoId
			for i := range undesiredVideoTitles {
				if strings.Contains(title, undesiredVideoTitles[i]) {
					title = strings.Replace(title, undesiredVideoTitles[i], "", -1)
				} else {
					continue
				}
			}
			playlist = append(playlist, title)
			fmt.Printf("%v, (%v)\r\n", title, videoId)
		}

		// Set the token to retrieve the next page of results
		// or exit the loop if all results have been retrieved.
		nextPageToken = playlistResponse.NextPageToken
		if nextPageToken == "" {
			break
		}
		fmt.Println()
	}

	return playlist
}

type Spotify struct {
}

func NewSpotify() Playlist {
	return &Spotify{}
}
func (S *Spotify) GetSongs() []string {
	var spotifyID spotify.ID
	client := getSpotifyClient(spotify.ScopeUserReadPrivate)
	service := spotify.NewClient(client)
	playlistId := playlistIDFromURL("spotify")
	spotifyID = spotify.ID(playlistId)
	playlist := spotifyPlaylistItems(service, spotifyID)
	fmt.Println(playlist) //placeholder for testing
	return playlist
}
