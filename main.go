package main

import (
	"fmt"
)

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
	fmt.Println("Retrieved spotify playlist")
	playlist := make(map[string]string)
	return playlist
}

func createSpotifyPlaylist(playlist map[string]string) { //creates spotify playlist from the text file that is a playlist
	fmt.Println("Added songs to apple music")
}

func getYouTubePlaylist() map[string]string { //gets YouTube playlist and writes it to a text file

	fmt.Println("Retrieved Youtube playlist")
	playlist := make(map[string]string)
	return playlist
}
func createYouTubePlaylist(playlist map[string]string) {
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

	//Ask what they are converting from, and assign to start
	start, finish = determineFlow()
	for {
		if start == finish {
			fmt.Println("Please make sure your start and ending services are different")
			start, finish = determineFlow()
			continue
		} else if start != 1 && start != 2 && start != 3 {
			fmt.Println("Please select a provided option.")
			start, finish = determineFlow()
			continue
		} else if finish != 1 && finish != 2 && finish != 3 {
			fmt.Println("Please select a provided option.")
			start, finish = determineFlow()
			continue
		} else {
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
	switch finish {
	case 1:
		createSpotifyPlaylist(playlist)
	case 2:
		createYouTubePlaylist(playlist)
	case 3:
		createApplePlaylist(playlist)
	}
	//this for loop prints out all the songs that couldn't be cloned

}
