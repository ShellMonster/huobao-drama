package cache

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/dgraph-io/ristretto"
)

type Entry struct {
	Data interface{}
	ETag string
}

var (
	cacheOnce sync.Once
	store     *ristretto.Cache
)

var (
	namespaceMu   sync.RWMutex
	namespaceVers = map[string]uint64{}
)

func getStore() *ristretto.Cache {
	cacheOnce.Do(func() {
		c, err := ristretto.NewCache(&ristretto.Config{
			NumCounters: 1e4,
			MaxCost:     64 << 20, // 64MB
			BufferItems: 64,
		})
		if err != nil {
			panic(fmt.Errorf("failed to init cache: %w", err))
		}
		store = c
	})
	return store
}

func Get(key string) (Entry, bool) {
	value, ok := getStore().Get(key)
	if !ok {
		return Entry{}, false
	}
	entry, ok := value.(Entry)
	return entry, ok
}

func Set(key string, data interface{}, ttl time.Duration) (Entry, error) {
	etag, cost, err := ComputeETag(data)
	if err != nil {
		return Entry{}, err
	}
	if cost < 1 {
		cost = 1
	}
	entry := Entry{Data: data, ETag: etag}
	getStore().SetWithTTL(key, entry, cost, ttl)
	return entry, nil
}

func Delete(key string) {
	getStore().Del(key)
}

func NamespaceKey(namespace string, key string) string {
	return fmt.Sprintf("%s:v%d:%s", namespace, getNamespaceVersion(namespace), key)
}

func NamespaceKeyWithQuery(namespace string, values url.Values) string {
	queryKey := buildQueryKey(values)
	if queryKey == "" {
		return NamespaceKey(namespace, "all")
	}
	return NamespaceKey(namespace, queryKey)
}

func BumpNamespace(namespace string) {
	namespaceMu.Lock()
	defer namespaceMu.Unlock()
	namespaceVers[namespace] = namespaceVers[namespace] + 1
	if namespaceVers[namespace] == 0 {
		namespaceVers[namespace] = 1
	}
}

func getNamespaceVersion(namespace string) uint64 {
	namespaceMu.RLock()
	version, ok := namespaceVers[namespace]
	namespaceMu.RUnlock()
	if ok && version > 0 {
		return version
	}

	namespaceMu.Lock()
	defer namespaceMu.Unlock()
	if version, ok := namespaceVers[namespace]; ok && version > 0 {
		return version
	}
	namespaceVers[namespace] = 1
	return 1
}

func buildQueryKey(values url.Values) string {
	if len(values) == 0 {
		return ""
	}
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var builder strings.Builder
	for _, key := range keys {
		vals := values[key]
		if len(vals) == 0 {
			continue
		}
		sort.Strings(vals)
		for _, val := range vals {
			if builder.Len() > 0 {
				builder.WriteByte('&')
			}
			builder.WriteString(url.QueryEscape(key))
			builder.WriteByte('=')
			builder.WriteString(url.QueryEscape(val))
		}
	}
	return builder.String()
}

func ComputeETag(data interface{}) (string, int64, error) {
	payload, err := json.Marshal(data)
	if err != nil {
		return "", 0, err
	}
	sum := sha256.Sum256(payload)
	return fmt.Sprintf("\"%x\"", sum), int64(len(payload)), nil
}

func MatchETag(ifNoneMatch string, etag string) bool {
	if ifNoneMatch == "" || etag == "" {
		return false
	}
	for _, token := range strings.Split(ifNoneMatch, ",") {
		trimmed := strings.TrimSpace(token)
		if strings.HasPrefix(trimmed, "W/") {
			trimmed = strings.TrimPrefix(trimmed, "W/")
		}
		if trimmed == etag {
			return true
		}
	}
	return false
}
