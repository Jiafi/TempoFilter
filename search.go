package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strconv"

	"github.com/zmb3/spotify/v2"
)

type MyTrack struct {
	Name  string
	Tempo float64
	Key   spotify.Key
	ID    spotify.ID
}

func main() {
	playlistPtr := flag.String("playlist", "", "Enter the name of the playlist you want to filter on.")
	tempoPtr := flag.Float64("tempo", 120, "Enter the tempo you want to filter on.")
	createFilteredPlaylistPtr := flag.Bool("create", false, "Set if you wan to create a playlist with the filtered songs.")
	flag.Parse()

	client := StartServer()

	// search for playlists containing "holiday"
	playlist := Search(client, playlistPtr)
	fmt.Println("   ", playlist.Name, playlist.Tracks)
	tracks, err := PlaylistTracks(client, playlist.ID)

	if err != nil {
		log.Fatalf("Failed to get playlist tracks: %v", err)
	}

	myTracks := TracksAudioAnalysis(client, tracks)

	filteredTracks := FilterTracks(myTracks, tempoPtr, nil)

	for _, track := range *filteredTracks {
		fmt.Println("  ", track.Name, track.Tempo)
	}

	if *createFilteredPlaylistPtr {
		newPlaylistName := playlist.Name + strconv.FormatFloat(*tempoPtr, 'f', 2, 32)
		CreatePlaylistWithFilteredTracks(client, newPlaylistName, filteredTracks)
	}
}

func CreatePlaylistWithFilteredTracks(client *spotify.Client, newPlaylistName string, tracks *[]MyTrack) {

	ctx := context.Background()
	currentUser, err := client.CurrentUser(ctx)
	if err != nil {
		log.Fatalf("There was an issue getting the current user: %v", err)
	}
	newPlaylist, err := client.CreatePlaylistForUser(ctx, currentUser.ID, newPlaylistName, "asdf", false, false)
	if err != nil {
		log.Fatalf("There was an issue creating the playlist: %v", err)
	}
	var filderedTrackIds []spotify.ID
	for _, track := range *tracks {
		filderedTrackIds = append(filderedTrackIds, track.ID)
	}

	client.AddTracksToPlaylist(ctx, newPlaylist.ID, filderedTrackIds...)
	log.Printf("Added Tracks to new playlist")
}

func Search(client *spotify.Client, name *string) *spotify.SimplePlaylist {
	ctx := context.Background()
	results, err := client.Search(ctx, *name, spotify.SearchTypePlaylist)

	if err != nil {
		log.Fatal(err)
	}

	if results.Playlists != nil {
		playlist := results.Playlists.Playlists[0]
		return &playlist
	}

	return nil
}

func PlaylistTracks(client *spotify.Client, playlistID spotify.ID) (*[]spotify.FullTrack, error) {
	ctx := context.Background()
	trackPage, err := client.GetPlaylistTracks(ctx, playlistID)
	if err != nil {
		return nil, err
	}

	var tracks []spotify.FullTrack

	for _, track := range trackPage.Tracks {
		tracks = append(tracks, track.Track)
	}
	return &tracks, err
}

func TracksAudioAnalysis(client *spotify.Client, tracks *[]spotify.FullTrack) *[]MyTrack {
	ctx := context.Background()
	var trackAnalysis []MyTrack
	for _, track := range *tracks {
		analysis, err := client.GetAudioAnalysis(ctx, track.ID)
		if err == nil {
			myTrack := MyTrack{Name: track.Name, Tempo: analysis.Track.Tempo, Key: analysis.Track.Key, ID: track.ID}
			trackAnalysis = append(trackAnalysis, myTrack)
		}
	}
	return &trackAnalysis
}

func FilterTracks(tracks *[]MyTrack, Tempo *float64, Key *spotify.Key) *[]MyTrack {
	tenPercentOfTempo := *Tempo * 0.1
	minTempo := *Tempo - tenPercentOfTempo
	maxTempo := *Tempo + tenPercentOfTempo

	var filteredTracks []MyTrack

	for _, track := range *tracks {
		check := true
		if Key != nil {
			check = check && track.Key == *Key
		}
		check = check && (minTempo <= track.Tempo && track.Tempo <= maxTempo)
		if check {
			filteredTracks = append(filteredTracks, track)
		}
	}
	return &filteredTracks
}
