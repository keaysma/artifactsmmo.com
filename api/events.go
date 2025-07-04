package api

import (
	"fmt"
	"time"

	"artifactsmmo.com/m/types"
)

type EventsResponse = []types.EventDetails

type ActiveEventsResponse = []types.ActiveEventDetails

func GetAllEvents(itype string, page int, size int) (*EventsResponse, error) {
	payload := map[string]string{
		"page": fmt.Sprintf("%d", page),
		"size": fmt.Sprintf("%d", size),
	}

	if itype != "" {
		payload["type"] = itype
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

type CachedProto struct {
	Data  *ActiveEventsResponse
	Epoch int64
}

var cacheGetAllActiveEvents *CachedProto = nil

func GetAllActiveEvents(page int, size int) (*ActiveEventsResponse, error) {
	epochNow := time.Now().Unix()
	if cacheGetAllActiveEvents != nil && epochNow-cacheGetAllActiveEvents.Epoch < 3 {
		return cacheGetAllActiveEvents.Data, nil
	}

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

	newCache := CachedProto{
		Epoch: epochNow,
		Data:  &out,
	}
	cacheGetAllActiveEvents = &newCache

	return &out, nil
}
