package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

var (
	update     bool
	calendarID string
)

func main() {
	flag.BoolVar(&update, "update", false, "update token")
	flag.StringVar(&calendarID, "calendar", "", "calendar id")
	credentialFile := flag.String("credential", "credential.json", "service-account json file")
	flag.Parse()
	ctx := context.Background()

	authConfigJson, err := os.ReadFile(*credentialFile)
	if err != nil {
		log.Fatalf("unable marshal credentials: %v", err)
	}

	srv, err := calendar.NewService(ctx, option.WithCredentialsJSON(authConfigJson))
	if err != nil {
		log.Fatalf("Unable to retrieve Calendar client: %v", err)
	}

	t := time.Now().Round(time.Hour)
	tMax := t.Add(time.Hour)
	events, err := srv.Events.List(calendarID).ShowDeleted(false).
		SingleEvents(true).TimeMin(t.Format(time.RFC3339)).TimeMax(tMax.Format(time.RFC3339)).MaxResults(10).OrderBy("startTime").Do()
	if err != nil {
		log.Fatalf("Unable to retrieve next ten of the user's events: %v", err)
	}
	fmt.Println("Upcoming events:")
	if len(events.Items) == 0 {
		log.Fatal("No upcoming events found.")
	} else {
		for _, item := range events.Items {
			date := item.Start.DateTime
			if date == "" {
				date = item.Start.Date
			}
			fmt.Printf("%v (%v) %v\n", date, item.Summary, item.Description)
		}
	}
}
