package main

import (
	"encoding/json"
	"fmt"
	"golang.org/x/oauth2/google"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path/filepath"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"google.golang.org/api/youtube/v3"
)

/*const missingClientSecretsMessage = `
Please configure OAuth 2.0
`*/
func handleError(err error, message string) {
	if message == "" {
		message = "Error making API call"
	}
	if err != nil {
		log.Fatalf(message+": %v", err.Error())
	}
}
func saveGoogleToken(file string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", file)
	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}
func googleTokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	t := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(t)
	defer f.Close()
	return t, err
}
func googleTokenCacheFile() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	tokenCacheDir := filepath.Join(usr.HomeDir, ".credentials")
	os.MkdirAll(tokenCacheDir, 0700)
	return filepath.Join(tokenCacheDir,
		url.QueryEscape("youtube-go-quickstart.json")), err
}
func getGoogleTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var code string
	if _, err := fmt.Scan(&code); err != nil {
		log.Fatalf("Unable to read authorization code %v", err)
	}

	tok, err := config.Exchange(oauth2.NoContext, code)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web %v", err)
	}
	return tok
}
func getGoogleClient(ctx context.Context, config *oauth2.Config) *http.Client {
	cacheFile, err := googleTokenCacheFile()
	if err != nil {
		log.Fatalf("Unable to get path to cached credential file. %v", err)
	}
	tok, err := googleTokenFromFile(cacheFile)
	if err != nil {
		tok = getGoogleTokenFromWeb(config)
		saveGoogleToken(cacheFile, tok)
	}
	return config.Client(ctx, tok)
}

func determineFlow() (int, int) { //gets user input for program flow
	var start int
	var finish int
	chooser := []string{"Spotify", "YouTube", "Apple Music"}
	fmt.Println("1. Spotify | 2. YouTube | 3. Apple Music")
	fmt.Println("What are you converting from?")
	fmt.Scan(&start)
	//ask what they are converting to, and assign to finish
	fmt.Println("What are you converting to?")
	fmt.Scan(&finish)
	fmt.Println("You are converting from", chooser[start-1], "to", chooser[finish-1])
	fmt.Println(start, finish)
	return start, finish
}

func getSpotifyPlaylist() map[string]string { //gets spotify playlist and writes it to a text file
	fmt.Println("Retrieved spotify playlist") //placeholder for testing
	playlist := make(map[string]string)
	return playlist
}

func createSpotifyPlaylist(playlist map[string]string) { //creates spotify playlist from the text file that is a playlist
	fmt.Println("Added songs to apple music")
}

func getYouTubePlaylist() map[string]string { //gets YouTube playlist and writes it to a text file
	ctx := context.Background()

	b, err := ioutil.ReadFile("client_secret.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved credentials
	// at ~/.credentials/youtube-go-quickstart.json
	config, err := google.ConfigFromJSON(b, youtube.YoutubeReadonlyScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getGoogleClient(ctx, config)
	service, err := youtube.New(client)

	handleError(err, "Error creating YouTube client")
	playlist := make(map[string]string)
	return playlist
}

func createYouTubePlaylist(playlist map[string]string) {
	ctx := context.Background()

	b, err := ioutil.ReadFile("client_secret.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved credentials
	// at ~/.credentials/youtube-go-quickstart.json
	config, err := google.ConfigFromJSON(b, youtube.YoutubepartnerScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getGoogleClient(ctx, config)
	service, err := youtube.New(client)

	handleError(err, "Error creating YouTube client")
	fmt.Println("created youtube playlist")
}

func getApplePlaylist() map[string]string { //retrieves Apple Music playlist and stores it as a text file
	fmt.Println("Retrieved Apple playlist")
	playlist := make(map[string]string)
	return playlist
}

func createApplePlaylist(playlist map[string]string) { //creates Apple Music playlist from the playlist file
	fmt.Println("Added songs to apple music")
}
func main() {
	var start int
	var finish int
	playlist := make(map[string]string)

	start, finish = determineFlow() //Ask what they are converting to and from, and assign to start and finish
	for {                           //checks that available options were selected and that start and finish are different. If either is false loop until both are true.
		if start == finish { //checks if start and finish are the same
			fmt.Println("Please make sure your start and ending services are different")
			start, finish = determineFlow()
			continue
		} else if start != 1 && start != 2 && start != 3 { // check start to make sure it is an available option
			fmt.Println("Please select a provided option.")
			start, finish = determineFlow()
			continue
		} else if finish != 1 && finish != 2 && finish != 3 { // check finish to make sure it is an available option
			fmt.Println("Please select a provided option.")
			start, finish = determineFlow()
			continue
		} else { //break out of loop once all conditions satisfied
			break
		}
	}
	//start getting the playlist
	switch start {
	case 1:
		playlist = getSpotifyPlaylist()
	case 2:
		playlist = getYouTubePlaylist()
	case 3:
		playlist = getApplePlaylist()
	}
	//copy playlist to other service
	switch finish {
	case 1:
		createSpotifyPlaylist(playlist)
	case 2:
		createYouTubePlaylist(playlist)
	case 3:
		createApplePlaylist(playlist)
	}
	//print failed songs to text file.

}
