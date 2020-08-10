package stats

import (
	"bytes"
	"flag"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"
)

var printStats = flag.Bool("stats", false, "Print stats to console")

func IncCounter(name string, tags map[string]string, value int64) {
	name = addTagsToName(name, tags)
	// case3 : todo 优化buf
	//name = addTagsToNameFast(name, tags)
	if *printStats {
		fmt.Printf("IncCounter: %v = %v\n", name, value)
	}
}

func UpdateGauge(name string, tags map[string]string, value int64) {
	name = addTagsToName(name, tags)
	// case3 : todo 优化buf
	//name = addTagsToNameFast(name, tags)
	if *printStats {
		fmt.Printf("UpdateGauge: %v = %v\n", name, value)
	}
}

func RecordTimer(name string, tags map[string]string, d time.Duration) {
	name = addTagsToName(name, tags)
	// case3 : todo 优化buf
	//name = addTagsToNameFast(name, tags)
	if *printStats {
		fmt.Printf("RecordTimer: %v = %v\n", name, d)
	}
}

var bufPool = sync.Pool{
	New: func() interface{} {
		return &bytes.Buffer{}
	},
}

func addTagsToName(name string, tags map[string]string) string {

	var keyOrder []string
	if _, ok := tags["host"]; ok {
		keyOrder = append(keyOrder, "host")
	}
	keyOrder = append(keyOrder, "endpoint", "os", "browser")

	parts := []string{name}
	for _, k := range keyOrder {
		v, ok := tags[k]
		if !ok || v == "" {
			parts = append(parts, "no-"+k)
			continue
		}
		parts = append(parts, clean(v))
		// case 2: todo remove reg
		//parts = append(parts, clean0(v))
	}

	return strings.Join(parts, ".")
}

var specialChars = regexp.MustCompile(`[{}/\\:\s.]`)

// clean takes a string that may contain special characters, and replaces these
// characters with a '-'.
func clean(value string) string {
	return specialChars.ReplaceAllString(value, "-")
}

func clean0(value string) string {
	newStr := make([]byte, len(value))
	for i := 0; i < len(value); i++ {
		switch c := value[i]; c {
		case '{', '}', '/', '\\', ':', ' ', '\t', '.':
			newStr[i] = '-'
		default:
			newStr[i] = c
		}
	}
	return string(newStr)
}

// here is optimized.

func addTagsToNameFast(name string, tags map[string]string) string {
	// The format we want is: host.endpoint.os.browser
	// if there's no host tag, then we don't use it.
	keyOrder := make([]string, 0, 4)
	if _, ok := tags["host"]; ok {
		keyOrder = append(keyOrder, "host")
	}
	keyOrder = append(keyOrder, "endpoint", "os", "browser")

	// We tried to pool the object, but perf didn't get better.
	// It's most likely due to use of defer, which itself has non-trivial overhead.
	// buf := &bytes.Buffer{}
	buf := bufPool.Get().(*bytes.Buffer)
	defer bufPool.Put(buf)
	buf.Reset()

	buf.WriteString(name)
	for _, k := range keyOrder {
		buf.WriteByte('.')

		v, ok := tags[k]
		if !ok || v == "" {
			buf.WriteString("no-")
			buf.WriteString(k)
			continue
		}

		writeClean(buf, v)
	}

	return buf.String()
}

func writeClean(buf *bytes.Buffer, value string) {
	for i := 0; i < len(value); i++ {
		switch c := value[i]; c {
		case '{', '}', '/', '\\', ':', ' ', '\t', '.':
			buf.WriteByte('-')
		default:
			buf.WriteByte(c)
		}
	}
}
