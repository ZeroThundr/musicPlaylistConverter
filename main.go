package main

import (
	"encoding/json"
	"fmt"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	_ "golang.org/x/oauth2/spotify"
	"google.golang.org/api/youtube/v3"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
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

// This variable indicates whether the script should launch a web server to
// initiate the authorization flow or just display the URL in the terminal
// window. Note the following instructions based on this setting:
// * launchWebServer = true
//   1. Use OAuth2 credentials for a web application
//   2. Define authorized redirect URIs for the credential in the Google APIs
//      Console and set the RedirectURL property on the config object to one
//      of those redirect URIs. For example:
//      config.RedirectURL = "http://localhost:8080"
//   3. In the startWebServer function below, update the URL in this line
//      to match the redirect URI you selected:
//         listener, err := net.Listen("tcp", "localhost:8080")
//      The redirect URI identifies the URI to which the user is sent after
//      completing the authorization flow. The listener then captures the
//      authorization code in the URL and passes it back to this script.
// * launchWebServer = false
//   1. Use OAuth2 credentials for an installed application. (When choosing
//      the application type for the OAuth2 client ID, select "Other".)
//   2. Set the redirect URI to "urn:ietf:wg:oauth:2.0:oob", like this:
//      config.RedirectURL = "urn:ietf:wg:oauth:2.0:oob"
//   3. When running the script, complete the auth flow. Then copy the
//      authorization code from the browser and enter it on the command line.
const launchWebServer = true

const missingClientSecretsMessage = `
Please configure OAuth 2.0
To make this sample run, you need to populate the client_secrets.json file
found at:
   %v
with information from the {{ Google Cloud Console }}
{{ https://cloud.google.com/console }}
For more information about the client_secrets.json file format, please visit:
https://developers.google.com/api-client-library/python/guide/aaa_client_secrets
`

// getClient uses a Context and Config to retrieve a Token
// then generate a Client. It returns the generated Client.
func getGoogleClient(scope string) *http.Client {
	ctx := context.Background()

	b, err := ioutil.ReadFile("googleClientSecret.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
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

	cacheFile, err := tokenCacheFile()
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

// startWebServer starts a web server that listens on http://localhost:8080.
// The webserver waits for an oauth code in the three-legged auth flow.
func startWebServer() (codeCh chan string, err error) {
	listener, err := net.Listen("tcp", "localhost:8080")
	if err != nil {
		return nil, err
	}
	codeCh = make(chan string)

	go http.Serve(listener, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		code := r.FormValue("code")
		codeCh <- code // send code to OAuth flow
		listener.Close()
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "Received code: %v\r\nYou can now safely close this browser window.", code)
	}))

	return codeCh, nil
}

// openURL opens a browser window to the specified location.
// This code originally appeared at:
//   http://stackoverflow.com/questions/10377243/how-can-i-launch-a-process-that-is-not-a-file-in-go
func openURL(url string) error {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", "http://localhost:4001/").Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("Cannot open URL %s on this platform", url)
	}
	return err
}

// Exchange the authorization code for access token
func exchangeToken(config *oauth2.Config, code string) (*oauth2.Token, error) {
	tok, err := config.Exchange(oauth2.NoContext, code)
	if err != nil {
		log.Fatalf("Unable to retrieve token %v", err)
	}
	return tok, nil
}

// getTokenFromPrompt uses Config to request a Token and prompts the user
// to enter the token on the command line. It returns the retrieved Token.
func getTokenFromPrompt(config *oauth2.Config, authURL string) (*oauth2.Token, error) {
	var code string
	fmt.Printf("Go to the following link in your browser. After completing "+
		"the authorization flow, enter the authorization code on the command "+
		"line: \n%v\n", authURL)

	if _, err := fmt.Scan(&code); err != nil {
		log.Fatalf("Unable to read authorization code %v", err)
	}
	fmt.Println(authURL)
	return exchangeToken(config, code)
}

// getTokenFromWeb uses Config to request a Token.
// It returns the retrieved Token.
func getTokenFromWeb(config *oauth2.Config, authURL string) (*oauth2.Token, error) {
	codeCh, err := startWebServer()
	if err != nil {
		fmt.Printf("Unable to start a web server.")
		return nil, err
	}

	err = openURL(authURL)
	if err != nil {
		log.Fatalf("Unable to open authorization URL in web server: %v", err)
	} else {
		fmt.Println("Your browser has been opened to an authorization URL.",
			" This program will resume once authorization has been provided.\n")
		fmt.Println(authURL)
	}

	// Wait for the web server to get the code.
	code := <-codeCh
	return exchangeToken(config, code)
}

// tokenCacheFile generates credential file path/filename.
// It returns the generated credential path/filename.
func tokenCacheFile() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	tokenCacheDir := filepath.Join(usr.HomeDir, ".credentials")
	os.MkdirAll(tokenCacheDir, 0700)
	return filepath.Join(tokenCacheDir,
		url.QueryEscape("youtube-go.json")), err
}

// tokenFromFile retrieves a Token from a given file path.
// It returns the retrieved Token and any read error encountered.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	t := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(t)
	defer f.Close()
	return t, err
}

// saveToken uses a file path to create a file and store the
// token in it.
func saveToken(file string, token *oauth2.Token) {
	fmt.Println("trying to save token")
	fmt.Printf("Saving credential file to: %s\n", file)
	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func playlistItemsList(service *youtube.Service, part []string, playlistId string, pageToken string) *youtube.PlaylistItemListResponse {
	call := service.PlaylistItems.List(part)
	call = call.PlaylistId(playlistId)
	if pageToken != "" {
		call = call.PageToken(pageToken)
	}
	response, err := call.Do()
	handleError(err, "")
	return response
}

func determineFlow() (int, int) { //gets user input for program flow
	var start int
	var finish int
	chooser := []string{"Spotify", "YouTube" /*, "Apple Music"*/}
	fmt.Println("1. Spotify " + "| 2. YouTube " /*+"| 3. Apple Music"*/)
	fmt.Println("What are you converting from?")
	fmt.Scan(&start)
	//ask what they are converting to, and assign to finish
	fmt.Println("What are you converting to?")
	fmt.Scan(&finish)
	fmt.Println("You are converting from", chooser[start-1], "to", chooser[finish-1])
	fmt.Println(start, finish)
	return start, finish
}

func getSpotifyPlaylist() []string { //gets spotify playlist and writes it to a text file
	playlistLength := 5
	fmt.Println("Retrieved spotify playlist") //placeholder for testing
	playlist := make([]string, playlistLength)
	return playlist
}

func createSpotifyPlaylist(playlist []string) { //creates spotify playlist from the text file that is a playlist
	fmt.Println("Added songs to Spotify")
}

func getYouTubePlaylist() []string { //gets YouTube playlist and writes it to a text file
	//ctx := context.Background()
	playlist := make([]string, 0)
	part := []string{"snippet"}
	/*b, err := ioutil.ReadFile("client_secret.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}*/
	client := getGoogleClient(youtube.YoutubeReadonlyScope)
	service, err := youtube.New(client)

	if err != nil {
		log.Fatalf("Error creating YouTube client: %v", err)
	}

	var playlistId = "PL1q3UCTMFFnrNb236IePHLSzHl7oQGOTo" // Print the playlist ID for the list of uploaded videos.
	fmt.Printf("Videos in list %s\r\n", playlistId)

	nextPageToken := ""
	for {
		// Retrieve next set of items in the playlist.
		playlistResponse := playlistItemsList(service, part, playlistId, nextPageToken)

		for _, playlistItem := range playlistResponse.Items {
			title := playlistItem.Snippet.Title
			videoId := playlistItem.Snippet.ResourceId.VideoId
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

func createYouTubePlaylist(playlist []string) {
	/*ctx := context.Background()
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
	client := getGoogleClient(youtube.YoutubepartnerScope)
	service, err := youtube.NewService(ctx)
	handleError(err, "Error creating YouTube client")*/
	fmt.Println("created youtube playlist")
	return
}
func main() {
	var start int
	var finish int
	var playlist []string

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
	}
	//copy playlist to other service
	switch finish {
	case 1:
		createSpotifyPlaylist(playlist)
		println(playlist[1])
	case 2:
		createYouTubePlaylist(playlist)
		println(playlist[1])

	}
}
