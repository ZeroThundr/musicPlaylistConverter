package main

import (
	"encoding/json"
	"fmt"
	"golang.org/x/oauth2"
	_ "golang.org/x/oauth2/spotify"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"runtime"
)

type Service int

const (
	SPOTIFY Service = iota
	YOUTUBE
)

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
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
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
			" This program will resume once authorization has been provided.")
		fmt.Println(authURL)
	}

	// Wait for the web server to get the code.
	code := <-codeCh
	return exchangeToken(config, code)
}

// tokenCacheFile generates credential file path/filename.
// It returns the generated credential path/filename.
func tokenCacheFile(clientName string) (string, error) {
	var fileName = clientName + "-go.json"
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	tokenCacheDir := filepath.Join(usr.HomeDir, ".credentials")
	os.MkdirAll(tokenCacheDir, 0700)
	return filepath.Join(tokenCacheDir,
		url.QueryEscape(fileName)), err
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
func playlistIDFromURL(service string) string { //extracts playlist id from url based on which service is being used
	//only returns the ID portion of the URLS
	var playlistURL string
	var playlistID string
	switch service {
	case "spotify":
		var re = regexp.MustCompile(`(?P<spotifyURL>\Qhttps://open.spotify.com/playlist/\E)(?P<id>[A-Za-z0-9]{22})`) //spotify playlist URL format
		fmt.Println("Enter the Spotify playlist URL")
		_, err := fmt.Scan(&playlistURL)
		if err != nil {
			fmt.Println("Please try again")
			playlistIDFromURL("spotify")
		}
		if re.MatchString(playlistURL) {
			matches := re.FindStringSubmatch(playlistURL)
			indexID := re.SubexpIndex("id")
			playlistID = matches[indexID]
		} else {
			fmt.Println("Please try again")
			playlistIDFromURL("spotify")
		}
	case "youtube":
		var re = regexp.MustCompile(`(?m)(?P<youtubeurl>\Qhttps://www.youtube.com/playlist?list=\E)(?P<id>.{34})`) //youtube URL format
		fmt.Println("Enter the YouTube playlist URL")
		_, err := fmt.Scan(&playlistURL)
		if err != nil {
			fmt.Println("please try again")
			playlistIDFromURL("youtube")
		}
		if re.MatchString(playlistURL) {
			matches := re.FindStringSubmatch(playlistURL)
			indexID := re.SubexpIndex("id")
			playlistID = matches[indexID]
		} else {
			fmt.Println("Please input a valid YouTube playlist ID")
			playlistIDFromURL("youtube")
		}
	default:
		log.Fatalf("INVALID SERVICE")
	}
	return playlistID
}

func determineFlow() (Service, Service) { //gets user input for program flow
	var start Service
	var finish Service
	var err error
	chooser := []string{"Spotify", "YouTube" /*, "Apple Music"*/}
	fmt.Println("0. Spotify " + "| 1. YouTube " /*+"| 3. Apple Music"*/)
	fmt.Println("What are you converting from?")
	_, err = fmt.Scan(&start)
	if err != nil {
		fmt.Println("please try again.")
		determineFlow()
	}
	//ask what they are converting to, and assign to finish
	fmt.Println("What are you converting to?")
	_, err = fmt.Scan(&finish)
	if err != nil {
		fmt.Println("please try again.")
		determineFlow()
	}
	fmt.Println("You are converting from", chooser[start], "to", chooser[finish])
	fmt.Println(start, finish)
	return start, finish
}
func cleanUp() { //deletes credential files
	services := []string{"youtube", "spotify"}
	usr, err := user.Current()
	if err != nil {
		log.Fatalf("could not retrieve current user")
	}
	for i := range services {
		err = os.Remove(usr.HomeDir + "/.credentials/" + services[i] + "-go.json")
		if err != nil {
			log.Fatalf("error deleting file")
		}
	}
}
func main() {
	var start Service
	var finish Service
	var playlist Playlist

	start, finish = determineFlow() //Ask what they are converting to and from, and assign to start and finish
	for {                           //checks that available options were selected and that start and finish are different. If either is false loop until both are true.
		if start == finish { //checks if start and finish are the same
			fmt.Println("Please make sure your start and ending services are different")
			start, finish = determineFlow()
			continue
		} else if start != SPOTIFY && start != YOUTUBE /* && start != 3 */ { // check start to make sure it is an available option
			fmt.Println("Please select a provided option.")
			start, finish = determineFlow()
			continue
		} else if finish != SPOTIFY && finish != YOUTUBE /* && finish != 3*/ { // check finish to make sure it is an available option
			fmt.Println("Please select a provided option.")
			start, finish = determineFlow()
			continue
		} else { //break out of loop once all conditions satisfied
			break
		}
	}
	//start getting the playlist
	switch start {
	case SPOTIFY:
		playlist = NewSpotify()
	case YOUTUBE:
		playlist = NewYoutube()
	}
	//copy playlist to other service
	switch finish {
	case SPOTIFY:
		createSpotifyPlaylist(playlist.GetSongs())
		fmt.Println("completed!")
	case YOUTUBE:
		createYouTubePlaylist(playlist.GetSongs())
		fmt.Println("Completed!")

	}
	cleanUp()
}
