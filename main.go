package main

import (
	"fmt"
)

/*func getSpotifyPlaylist() { //gets spotify playlist and writes it to a text file

}
func createSpotifyPlaylist() { //creates spotify playlist from the text file that is a playlist
}

func getYouTubePlaylist() { //gets YouTube playlist and writes it to a text file

}

func getApplePlaylist() { //retrieves apple music playlist and stores it as a text file

}

func createApplePlaylist() { //creates Apple Music playlist from the playlist file

}*/
//picks the names for the converter
func chooser(number int) string {
	switch number {
	case 1:
		return "Spotify"
	case 2:
		return "YouTube"
	case 3:
		return "Apple Music"
	default:
		return "ERR0R"
	}
}
func main() {
	var start int
	var finish int
	//Ask what they are converting from, and assign to start
	fmt.Println("1. Spotify | 2. YouTube | 3. Apple Music")
	fmt.Println("What are you converting from?")
	fmt.Scan(&start)
	//ask what they are converting to, and assign to finish
	fmt.Println("What are you converting to?")
	scan, err := fmt.Scan(&finish)
	if err != nil {
		return
	}
	fmt.Println("You are converting from", chooser(start), "to", chooser(finish))
	fmt.Println(start, finish)
}
