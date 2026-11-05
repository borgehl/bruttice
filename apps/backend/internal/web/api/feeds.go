package api

import (
	"fmt"
	"github.com/borgehl/bruttice/packages/go-pkg/models"
	"github.com/borgehl/bruttice/packages/go-pkg/platform/util"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/feeds"
)

func feedDescription(event models.Event) string {
	ret := fmt.Sprintf("%s at %s: ", event.StartAt.In(util.TZLocation).Format("3:04pm"), event.Venue.Name)
	if event.CoverCost != "" {
		ret = fmt.Sprintf("%s %s,", ret, event.CoverCost)
	}
	return fmt.Sprintf("%s %s", ret, event.Age.Name)
}

func feedEventTitle(event models.Event) string {
	title := event.Title
	if event.CancelledAt != nil {
		title = "CANCELLED: " + title
	}
	return fmt.Sprintf("%s: %s", event.StartAt.In(util.TZLocation).Format("Monday 01-02-06"), title)
}

func (h *APIHandler) handleGetFeeds(w http.ResponseWriter, r *http.Request) {
	tx := h.database.DB.Begin()
	defer tx.Commit()

	filter, err := h.getFilter(r)
	if err != nil {
		h.logger.Errorf("failed to get filter for feeds: %v", err)
		http.Error(w, "Failed to get filter", http.StatusInternalServerError)
		return
	}

	events, err := h.database.GetIndexEvents(filter, tx)
	if err != nil {
		h.logger.Errorf("failed to get events for feed: %v", err)
		http.Error(w, "Failed to get events", http.StatusInternalServerError)
		return
	}

	feed := &feeds.Feed{
		Title:       os.Getenv("FEED_TITLE"),
		Link:        &feeds.Link{Href: os.Getenv("HOST") + "/api/feeds"},
		Description: os.Getenv("FEED_DESCRIPTION"),
		Author:      &feeds.Author{Name: os.Getenv("FEED_AUTHOR")},
		Created:     time.Now(),
	}

	feed.Items = make([]*feeds.Item, 0, len(events))
	for _, event := range events {
		feed.Items = append(feed.Items, &feeds.Item{
			Title:       feedEventTitle(event),
			Id:          event.Guid,
			IsPermaLink: strconv.FormatBool(false),
			Link:        &feeds.Link{Href: event.Link},
			Description: feedDescription(event),
			Created:     event.StartAt,
			Updated:     event.StartAt,
		})
	}

	feedType := r.URL.Query().Get("feedType")
	switch feedType {
	case "", "rss":
		rss, err := feed.ToRss()
		if err != nil {
			h.logger.Errorf("failed to generate RSS feed: %v", err)
			http.Error(w, "Failed to generate feed", http.StatusInternalServerError)
			return
		}
		w.Header().Add("Content-Type", "text/xml;charset=UTF-8")
		w.Write([]byte(rss))

	case "atom":
		atom, err := feed.ToAtom()
		if err != nil {
			h.logger.Errorf("failed to generate Atom feed: %v", err)
			http.Error(w, "Failed to generate feed", http.StatusInternalServerError)
			return
		}
		w.Header().Add("Content-Type", "text/xml;charset=UTF-8")
		w.Write([]byte(atom))

	case "json":
		jsonFeed, err := feed.ToJSON()
		if err != nil {
			h.logger.Errorf("failed to generate JSON feed: %v", err)
			http.Error(w, "Failed to generate feed", http.StatusInternalServerError)
			return
		}
		w.Header().Add("Content-Type", "application/json")
		w.Write([]byte(jsonFeed))
	}
}
