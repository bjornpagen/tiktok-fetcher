package main

import (
	"fmt"
	"os"
	"sync"

	"github.com/bjornpagen/tiktok-video-processor/pkg/server"
)

const (
	localDB = "users_db"
	outDir  = "out"
)

var (
	scraperKey string
	fetcherKey string
)

func init() {
	// Set up environment variables
	scraperKey = os.Getenv("SCRAPER_KEY")
	fetcherKey = os.Getenv("FETCHER_KEY")

	if scraperKey == "" || fetcherKey == "" {
		panic("SCRAPER_KEY and FETCHER_KEY must be set")
	}
}

var subCommands = []string{"add", "update", "fetch"}

func formatSubcommands(s []string) string {
	var out string
	for _, c := range s {
		out += c + "|"
	}
	return out[:len(out)-1]
}

func usage() {
	println("usage: go run main.go [" + formatSubcommands(subCommands) + "]")
}

func main() {
	if len(os.Args) < 2 {
		usage()
		return
	}

	s := server.New(localDB, outDir, scraperKey, fetcherKey)
	if err := s.DB.Open(); err != nil {
		panic(err)
	}
	defer s.DB.Close()

	var err error
	switch os.Args[1] {
	case "add":
		err = add(s)
	case "update":
		err = update(s)
	case "fetch":
		err = fetch(s)
	default:
		usage()
	}

	if err != nil {
		panic(err)
	}
}

func add(s *server.Server) error {
	if len(os.Args) < 3 {
		println("You need to specify a username")
		return fmt.Errorf("no username specified")
	}
	username := os.Args[2]
	if err := s.AddUsername(username); err != nil {
		return err
	}

	return nil
}

func update(s *server.Server) error {
	if err := s.UpdateAllOnce(); err != nil {
		return err
	}
	return nil
}

func fetch(s *server.Server) error {
	ids, err := s.DB.GetUserIDList()
	if err != nil {
		return err
	}

	// Above loop but concurrent.
	var wg sync.WaitGroup
	for _, id := range ids {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			if err := s.FetchAllVideos(id); err != nil {
				panic(err)
			}
		}(id)
	}
	wg.Wait()

	return nil
}
