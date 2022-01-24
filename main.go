package main

import (
	"fmt"
)

func determineFlow() (int, int) { //gets user input for program flow
	var start int
	var finish int
	fmt.Println("1. Spotify | 2. YouTube | 3. Apple Music")
	fmt.Println("What are you converting from?")
	fmt.Scan(&start)
	//ask what they are converting to, and assign to finish
	fmt.Println("What are you converting to?")
	fmt.Scan(&finish)
	fmt.Println("You are converting from", chooser(start), "to", chooser(finish))
	fmt.Println(start, finish)
	return start, finish
}

func getSpotifyPlaylist() { //gets spotify playlist and writes it to a text file
	fmt.Println("Retrieved spotify playlist")
}

func createSpotifyPlaylist() { //creates spotify playlist from the text file that is a playlist
	fmt.Println("Added songs to apple music")
}

func getYouTubePlaylist() { //gets YouTube playlist and writes it to a text file
	fmt.Println("Retrieved Youtube playlist")
}
func createYouTubePlaylist() {
	fmt.Println("created youtube playlist")
}

func getApplePlaylist() { //retrieves apple music playlist and stores it as a text file
	fmt.Println("Retrieved Apple playlist")
}

func createApplePlaylist() { //creates Apple Music playlist from the playlist file
	fmt.Println("Added songs to apple music")
}

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
		} else {
			break
		}
	}
	fmt.Println(start, finish)
	if start != finish {
		if start == 1 {
			getSpotifyPlaylist()
		} else if start == 2 {
			getYouTubePlaylist()
		} else if start == 3 {
			getApplePlaylist()
		} else {

		}
	} else {

	}

}
