package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/regclient/regclient"
	"github.com/regclient/regclient/types/ref"
)

type IRegistry interface {
	FetchImageTags(image string) ([]string, error)
}

type Registry struct {
	url string
}

func NewRegistry(url string) *Registry {
	return &Registry{url: url}
}

func (r *Registry) FetchImageTags(image string) ([]string, error) {
	if r.url == "" {
		return fetchImageTags(image)
	} else {
		return r.fetchImageTags(image)
	}
}

func fetchImageTags(image string) ([]string, error) {
	ctx := context.Background()
	rc := regclient.New()
	r, err := ref.New(image)
	if err != nil {
		return nil, fmt.Errorf("failed to create ref: %w", err)
	}
	defer rc.Close(ctx, r)
	tagList, err := rc.TagList(ctx, r)
	if err != nil {
		return nil, fmt.Errorf("failed to get tag list: %w", err)
	}
	return tagList.Tags, nil
}

func (r *Registry) fetchImageTags(image string) ([]string, error) {
	var tags []string
	url := r.getImageURL(image)

	for {
		resp, err := http.Get(url)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		type Tag struct {
			Name string `json:"name"`
		}

		var tagsResponse struct {
			Count   int    `json:"count"`
			Results []Tag  `json:"results"`
			Next    string `json:"next"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&tagsResponse); err != nil {
			return nil, err
		}

		for _, tag := range tagsResponse.Results {
			tags = append(tags, tag.Name)
		}

		if tagsResponse.Next == "" {
			break
		}

		url = tagsResponse.Next
	}

	return tags, nil
}

func (r *Registry) isOfficialImage(image string) bool {
	return strings.Count(image, "/") == 0
}

func (r *Registry) getImageURL(image string) string {
	// Remove tags from image name
	image = strings.Split(image, ":")[0]
	baseUrl := r.url
	if !strings.HasSuffix(baseUrl, "/") {
		baseUrl += "/"
	}
	if r.isOfficialImage(image) {
		baseUrl += "library/"
	}
	baseUrl += image + "/tags?page_size=100"
	return baseUrl
}
