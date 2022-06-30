package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/zmb3/spotify"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"io/ioutil"
	"log"
	"math"
	"net/http"
)

func SpotifyConfigFromJSON(jsonKey []byte, scope ...string) (*oauth2.Config, error) {
	type cred struct {
		ClientID     string   `json:"client_id"`
		ClientSecret string   `json:"client_secret"`
		RedirectURIs []string `json:"redirect_uris"`
		AuthURI      string   `json:"auth_uri"`
		TokenURI     string   `json:"token_uri"`
		ResponseType string   `json:"response_type"`
	}
	var j struct {
		Web       *cred `json:"web"`
		Installed *cred `json:"installed"`
	}
	if err := json.Unmarshal(jsonKey, &j); err != nil {
		return nil, err
	}
	var c *cred
	switch {
	case j.Web != nil:
		c = j.Web
	case j.Installed != nil:
		c = j.Installed
	default:
		return nil, fmt.Errorf("oauth2/spotify: no credentials found")
	}
	if len(c.RedirectURIs) < 1 {
		return nil, errors.New("oauth2/spotify: missing redirect URL in the client_credentials.json")
	}
	return &oauth2.Config{
		ClientID:     c.ClientID,
		ClientSecret: c.ClientSecret,
		RedirectURL:  c.RedirectURIs[0],
		Scopes:       scope,
		Endpoint: oauth2.Endpoint{
			AuthURL:  c.AuthURI,
			TokenURL: c.TokenURI,
		},
	}, nil
}
func getSpotifyClient(scope string) *http.Client {
	ctx := context.Background()
	b, err := ioutil.ReadFile("spotifyClientSecret.json")
	if err != nil {
		log.Fatalf("Unable to read spotify client secret file: %v", err)
	}
	config, err := SpotifyConfigFromJSON(b, scope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	// Use a redirect URI like this for a web app. The redirect URI must be a
	// valid one for your OAuth2 credentials.
	config.RedirectURL = "http://localhost:8080"
	// Use the following redirect URI if launchWebServer=false in oauth2.go
	// config.RedirectURL = "urn:ietf:wg:oauth:2.0:oob"

	cacheFile, err := tokenCacheFile("spotify")
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
func spotifyPlaylistItems(service spotify.Client, playlistId spotify.ID) []string { //gets list of spotify tracks in a playlist
	var spotifyPlaylistItemsList []string
	maxResult := 50
	itemsPage := 0
	options := spotify.Options{
		Limit:  &maxResult,
		Offset: &itemsPage,
	}
	playlistTracks, err := service.GetPlaylistTracksOpt(playlistId, &options, "total,items(track(name,artists))") //returns only the track name and artist
	if err != nil {
		log.Fatalf("Retrieving Playlist failed")
	}
	pages := int(math.Ceil(float64(playlistTracks.Total) / 50))

	if playlistTracks.Total > 50 { //Spotify can only grab 50 items at a time from a playlist, if the playlist is bigger than 50 items it iterates through multiple times
		i := 0
		x := 0
		for i <= pages {
			i++
			x = 0
			itemsPage = i * 50 //next page of results
			for x <= 49 {
				songInfo := playlistTracks.Tracks[x]
				songName := songInfo.Track.Name
				songArtist := songInfo.Track.Artists[0].Name
				searchQuery := songName + " - " + songArtist //concatenate song name and artist name for more accurate search
				spotifyPlaylistItemsList = append(spotifyPlaylistItemsList, searchQuery)
				x++
			}
		}
	} else {
		i := 0
		for i < playlistTracks.Total {
			songInfo := playlistTracks.Tracks[i]
			songName := songInfo.Track.Name
			songArtist := songInfo.Track.Artists[0].Name
			searchQuery := songName + " - " + songArtist
			fmt.Println(songName)
			spotifyPlaylistItemsList = append(spotifyPlaylistItemsList, searchQuery)
			i++
		}

	}
	return spotifyPlaylistItemsList
}
func getSpotifyPlaylist() []string { //gets spotify playlist and writes it to a text file
	var spotifyID spotify.ID
	client := getSpotifyClient(spotify.ScopeUserReadPrivate)
	service := spotify.NewClient(client)
	playlistId := playlistIDFromURL("spotify")
	spotifyID = spotify.ID(playlistId)
	playlist := spotifyPlaylistItems(service, spotifyID)
	fmt.Println(playlist) //placeholder for testing
	return playlist
}

func createSpotifyPlaylist(playlist []string) { //creates spotify playlist from the text file that is a playlist
	client := getSpotifyClient(spotify.ScopePlaylistModifyPrivate)
	service := spotify.NewClient(client)
	userInfo, err := service.CurrentUser()
	searchResultLimit := 1
	spotifyAddTrackLimit := 100
	options := spotify.Options{
		Limit: &searchResultLimit,
	}
	var spotifyTrackIds []spotify.ID
	if err != nil {
		log.Fatalf("Unable to retrieve user info")
	}
	userId := userInfo.ID
	playlistInfo, err := service.CreatePlaylistForUser(userId, "Converted Playlist", "", false)
	if err != nil {
		log.Fatalf("Unable to create playlist")
	}
	playlistId := playlistInfo.ID
	if len(playlist) > spotifyAddTrackLimit {
		x := 0
		for i := range playlist {
			if x < spotifyAddTrackLimit {
				searchResults, err := service.SearchOpt(playlist[i], spotify.SearchTypeTrack, &options)
				if err != nil {
					fmt.Printf("%s : not found/n", playlist[i])
					continue
				}
				songId := searchResults.Tracks.Tracks[0].ID
				spotifyTrackIds = append(spotifyTrackIds, songId)
				x += 1
			} else {
				snapshotId, err := service.AddTracksToPlaylist(playlistId, spotifyTrackIds...)
				if err != nil {
					log.Fatalf("Track ID:%s could not be added to Playlist ID: %s/n", spotifyTrackIds, playlistId)
				}
				fmt.Println(snapshotId)
				x = 0
			}
		}
	} else {
		for i := range playlist {
			searchResults, err := service.SearchOpt(playlist[i], spotify.SearchTypeTrack, &options)
			if err != nil {
				fmt.Printf("%s : not found", playlist[i])
				continue
			}
			songId := searchResults.Tracks.Tracks[0].ID
			spotifyTrackIds = append(spotifyTrackIds, songId)
		}
		snapshotId, err := service.AddTracksToPlaylist(playlistId, spotifyTrackIds...)
		if err != nil {
			log.Fatalf("Track ID:%s could not be added to Playlist ID: %s/n", spotifyTrackIds, playlistId)
		}
		fmt.Println(snapshotId)
	}
	fmt.Println("Added songs to Spotify")
}
