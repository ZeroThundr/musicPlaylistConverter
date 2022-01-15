package main

import (
	"fmt"
)

func getSpotifyPlaylist() { //gets spotify playlist and writes it to a text file

}
func createSpotifyPlaylist() { //creates spotify playlist from the text file that is a playlist
}

func getYouTubePlaylist() { //gets YouTube playlist and writes it to a text file

}

func getApplePlaylist() { //retrieves apple music playlist and stores it as a text file

}

func createApplePlaylist() { //creates apple music playlist from the playlist file

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
	fmt.Scan(&finish)
	fmt.Println(start, finish)
}
