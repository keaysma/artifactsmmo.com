package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"
	"sync"
	"time"

	"artifactsmmo.com/m/utils"
	"github.com/mitchellh/mapstructure"
	"github.com/redis/go-redis/v9"
)

type CacheEntry struct {
	Value string
	Epoch int
}

type SeqRedis struct {
	ctx   context.Context
	r     *redis.Client
	locke *sync.Mutex
}

var rx = SeqRedis{
	ctx: context.Background(),
	r: redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
		DB:   1,
	}),
	locke: &sync.Mutex{},
}

func (rdb SeqRedis) Get(key string) (string, error) {
	rdb.locke.Lock()
	defer rdb.locke.Unlock()
	return rdb.r.Get(rdb.ctx, key).Result()
}

func (rdb SeqRedis) Set(key string, value interface{}) (string, error) {
	rdb.locke.Lock()
	defer rdb.locke.Unlock()
	return rdb.r.Set(rdb.ctx, key, value, 0).Result()
}

var CACHE_CONFIG = map[string]time.Duration{
	"tasks/list/*":  24 * time.Hour,
	"events/active": 30 * time.Second,
	"events":        24 * time.Hour,
	"resources":     24 * time.Hour,
	"items/*":       24 * time.Hour,
	"items":         24 * time.Hour,
	"monsters/*":    24 * time.Hour,
	"monsters":      24 * time.Hour,
	"maps":          5 * time.Minute,
}

// structure:
/*
{
	  "data": {
		"name": "string",
		"skin": "men1",
	  }
}
*/

type Params *map[string]string
type Body interface{} //*map[string]interface{}

type DataResponse struct {
	Data interface{} `json:"data"`
}

func GetDataResponse[T interface{}](url string, params interface{}, response *T) error {
	hasCacheRule := false
	cacheKey := fmt.Sprintf("%s-%v", url, params)
	for rule, duration := range CACHE_CONFIG {
		regrule := strings.ReplaceAll(rule, "*", ".*")
		match, err := regexp.MatchString(
			fmt.Sprintf("^%s", regrule),
			url,
		)

		if err != nil {
			return err
		}

		if !match {
			continue
		}

		hasCacheRule = true
		utils.UniversalDebugLog(fmt.Sprintf("cache hit for %s", rule))

		entry, err := rx.Get(cacheKey)

		if err != nil {
			// Just go on if this doesn't work
			// Either there was no value, in which case this is the correct behavior
			// Or access to the cache is broken, but it's better the application just works in this case
			break
		}

		var cacheEntry CacheEntry
		err = json.Unmarshal([]byte(entry), &cacheEntry)

		if err != nil {
			return err
		}
		// utils.UniversalDebugLog(fmt.Sprintf("%s entry %v", rule, cacheEntry))

		durationSeconds := int(duration.Seconds())
		now := time.Now().Unix()
		// utils.UniversalDebugLog(fmt.Sprintf("entry.Epoch (%d) + duration (%d) = %d < %d ? %v then skip", cacheEntry.Epoch, durationSeconds, cacheEntry.Epoch+durationSeconds, int((now)), cacheEntry.Epoch+durationSeconds < int(now)))

		if cacheEntry.Epoch+durationSeconds < int(now) {
			break
		}

		var data DataResponse
		var uerr = json.Unmarshal([]byte(cacheEntry.Value), &data)

		if uerr != nil {
			return fmt.Errorf("failed to unmarshall JSON: %s, %s", err, cacheEntry.Value)
		}

		err = mapstructure.Decode(data.Data, response)
		if err != nil {
			return err
		}

		return nil
	}

	var parsedParams *map[string]string

	switch typedParams := params.(type) {
	case nil:
		parsedParams = &map[string]string{}
	case *map[string]string:
		parsedParams = typedParams
	default:
		utils.UniversalDebugLog("decoding interface")
		marshalledParamsRef, err := utils.MarshallParams(&typedParams)
		if err != nil {
			return err
		}

		utils.UniversalDebugLog(fmt.Sprintf("marshalled params: %s", marshalledParamsRef))
		parsedParams = marshalledParamsRef
	}

	res, err := utils.HttpGet(
		url,
		map[string]string{},
		parsedParams,
	)

	if err != nil {
		return err
	}

	text, err := io.ReadAll(res.Body)
	utils.UniversalDebugLog(fmt.Sprintf("response: %s", text))

	if err != nil {
		return err
	}

	var data DataResponse
	var uerr = json.Unmarshal(text, &data)

	if uerr != nil {
		return fmt.Errorf("failed to unmarshall JSON: %s, %s", err, text)
	}

	if data.Data == nil {
		return fmt.Errorf("missing data: %v, raw: %s", data, text)
	}

	err = mapstructure.Decode(data.Data, response)
	if err != nil {
		return err
	}

	if hasCacheRule {
		newCacheEntry := CacheEntry{
			Value: string(text),
			Epoch: int(time.Now().Unix()),
		}

		value, err := json.Marshal(newCacheEntry)
		if err != nil {
			return err
		}

		_, err = rx.Set(cacheKey, value)
		if err != nil {
			return err
		}
	}

	return nil
}

func PostDataResponse[T interface{}](url string, body Body, response *T) error {
	var rawBody []byte
	var err error

	if body != nil {
		rawBody, err = json.Marshal(body)
		if err != nil {
			return err
		}
	}

	res, err := utils.HttpPost(
		url,
		map[string]string{},
		bytes.NewReader(rawBody),
	)

	if err != nil {
		return err
	}

	sc := res.StatusCode
	if sc > 299 {
		text := ""
		b, err := io.ReadAll(res.Body)
		if err != nil {
			text = "unabled to decode!"
		} else {
			text = string(b)
		}

		return fmt.Errorf("error: %d, %s", sc, text)
	}

	text, err := io.ReadAll(res.Body)

	if err != nil {
		return err
	}

	var data DataResponse
	uerr := json.Unmarshal(text, &data)

	if uerr != nil {
		return err
	}

	err = mapstructure.Decode(data.Data, response)
	if err != nil {
		return err
	}

	return nil
}
