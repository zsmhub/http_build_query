package http_build_query

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/tidwall/gjson"
	"math"
	"net/url"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

type KVPair struct {
	Key, Value string
}

func HttpBuildQuery(data map[string]interface{}) (string, error) {
	ret := make(map[string]KVPair)
	keys := make([]string, 0)
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	padLen := int(math.Log10(float64(len(data)))) + 1
	for k, v := range keys {
		tempKey := v
		sortKey := padLeft(int64(k), padLen)

		switch reflect.TypeOf(data[tempKey]).Kind() {
		case reflect.Map:
			return "", errors.New("nested map[string]interface{} not supported")
		case reflect.Array, reflect.Slice, reflect.Struct:
			data, _ := json.Marshal(data[tempKey])
			httpBuildQueryInnerJson(string(data), tempKey, sortKey, ret)
		default:
			ret[sortKey] = KVPair{Key: tempKey, Value: fmt.Sprintf("%v", data[tempKey])}
		}
	}

	keys = make([]string, 0)
	for k := range ret {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	vals := make([]string, 0)
	for _, k := range keys {
		vals = append(vals, urlEncode(ret[k].Key)+"="+urlEncode(ret[k].Value))
	}

	return strings.Join(vals, "&"), nil
}

func HttpBuildQueryJson(jsonStr string) (string, error) {
	if IsJson(jsonStr) == false {
		return "", errors.New("param is not json string")
	}
	jsonData := gjson.Parse(jsonStr)
	if jsonData.IsObject() == false {
		return "", errors.New("param json must be a object")
	}

	tempMap := jsonData.Map()
	keys := make([]string, 0)
	for k := range tempMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	ret := make(map[string]KVPair)
	padLen := int(math.Log10(float64(len(keys)))) + 1

	for k, v := range keys {
		tempKey := v
		sortKey := padLeft(int64(k), padLen)

		tempJson := tempMap[v]
		if tempJson.IsArray() || tempJson.IsObject() {
			httpBuildQueryInnerJson(tempJson.Raw, tempKey, sortKey, ret)
		} else {
			val := tempJson.String()
			if val != "" {
				ret[sortKey] = KVPair{Key: tempKey, Value: val}
			}
		}
	}
	keys = make([]string, 0)
	for k := range ret {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	vals := make([]string, 0)
	for _, k := range keys {
		vals = append(vals, urlEncode(ret[k].Key)+"="+urlEncode(ret[k].Value))
	}
	return strings.Join(vals, "&"), nil
}

func httpBuildQueryInnerJson(json string, key string, sortk string, ret map[string]KVPair) {
	jsonData := gjson.Parse(json)

	if jsonData.IsObject() {
		tempMap := jsonData.Map()
		tempMapKeysIndex := make(map[string]int)
		tempMapKeys := make([]string, 0)
		for k := range tempMap {
			tempMapKeysIndex[k] = jsonData.Get(k).Index
			tempMapKeys = append(tempMapKeys, k)
		}
		sort.Slice(tempMapKeys, func(i, j int) bool {
			return tempMapKeysIndex[tempMapKeys[i]] < tempMapKeysIndex[tempMapKeys[j]]
		})

		padLen := int(math.Log10(float64(len(tempMapKeys)))) + 1
		for k, v := range tempMapKeys {
			sortKey := sortk + "." + padLeft(int64(k), padLen)
			tempKey := key + "[" + v + "]"

			tempJson := tempMap[v]
			if tempJson.IsArray() || tempJson.IsObject() {
				httpBuildQueryInnerJson(tempJson.Raw, tempKey, sortKey, ret)
			} else {
				val := tempJson.String()
				if val != "" {
					ret[sortKey] = KVPair{Key: tempKey, Value: val}
				}
			}
		}
	} else {
		tempMap := jsonData.Array()
		padLen := int(math.Log10(float64(len(tempMap)))) + 1
		for k, v := range tempMap {
			sortKey := sortk + "." + padLeft(int64(k), padLen)
			tempKey := key + "[" + strconv.Itoa(k) + "]"

			tempJson := v
			if tempJson.IsArray() || tempJson.IsObject() {
				httpBuildQueryInnerJson(tempJson.Raw, tempKey, sortKey, ret)
			} else {
				val := tempJson.String()
				if val != "" {
					ret[sortKey] = KVPair{Key: tempKey, Value: val}
				}
			}
		}
	}
}

func padLeft(v int64, length int) string {
	abs := math.Abs(float64(v))
	var padding int
	if v != 0 {
		min := math.Pow10(length - 1)

		if min-abs > 0 {
			l := math.Log10(abs)
			if l == float64(int64(l)) {
				l++
			}
			padding = length - int(math.Ceil(l))
		}
	} else {
		padding = length - 1
	}
	builder := strings.Builder{}
	if v < 0 {
		length = length + 1
	}
	builder.Grow(length * 4)
	if v < 0 {
		builder.WriteRune('-')
	}
	for i := 0; i < padding; i++ {
		builder.WriteRune('0')
	}
	builder.WriteString(strconv.FormatInt(int64(abs), 10))
	return builder.String()
}

func urlEncode(str string) string {
	specialChar := map[string]string{
		"~": "%7E",
	}
	for k, v := range specialChar {
		str = strings.Replace(url.QueryEscape(str), k, v, -1)
	}
	return str
}

func IsJson(s string) bool {
	var js map[string]interface{}
	return json.Unmarshal([]byte(s), &js) == nil
}
