package api

import (
	"fmt"

	"artifactsmmo.com/m/types"
)

type EventsResponse = []types.EventDetails

type ActiveEventsResponse = []types.ActiveEventDetails

func GetAllEvents(page int, size int) (*EventsResponse, error) {
	payload := map[string]string{
		"page": fmt.Sprintf("%d", page),
		"size": fmt.Sprintf("%d", size),
	}

	var out EventsResponse
	err := GetDataResponseFuture(
		"events",
		&payload,
		&out,
	)

	if err != nil {
		return nil, err
	}

	return &out, nil
}

func GetAllActiveEvents(page int, size int) (*ActiveEventsResponse, error) {
	payload := map[string]string{
		"page": fmt.Sprintf("%d", page),
		"size": fmt.Sprintf("%d", size),
	}

	var out ActiveEventsResponse
	err := GetDataResponseFuture(
		"events/active",
		&payload,
		&out,
	)

	if err != nil {
		return nil, err
	}

	return &out, nil
}
