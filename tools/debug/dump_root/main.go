package main

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	"github.com/yueli-fx/aep-parser/internal/rifx"
)

// parseTdsnName extracts the embedded Utf8 record name from a tdsn payload.
// Format: [ "Utf8" (4) | size uint32 BE (4) | name bytes ]
func parseTdsnName(d []byte) string {
	if len(d) < 8 || string(d[:4]) != "Utf8" {
		// Fall back to first NUL-terminated token
		n := bytes.IndexByte(d, 0)
		if n < 0 {
			n = len(d)
		}
		return string(d[:n])
	}
	size := binary.BigEndian.Uint32(d[4:8])
	if int(size) > len(d)-8 {
		size = uint32(len(d) - 8)
	}
	return string(d[8 : 8+size])
}

func main() {
	f, _ := os.Open(os.Args[1])
	defer f.Close()
	root, err := rifx.Parse(f)
	if err != nil {
		panic(err)
	}
	var dump func(c *rifx.Chunk, depth int)
	dump = func(c *rifx.Chunk, depth int) {
		ind := strings.Repeat("  ", depth)
		tag := string(c.ID[:])
		if c.IsList() {
			fmt.Printf("%sLIST %s (formType=%s, %d children)\n", ind, tag, string(c.FormType[:]), len(c.Children))
			for _, ch := range c.Children {
				dump(ch, depth+1)
			}
			return
		}
		if tag == "tdmn" {
			n := bytes.IndexByte(c.Data, 0)
			if n < 0 {
				n = len(c.Data)
			}
			fmt.Printf("%schunk %s name=%q\n", ind, tag, string(c.Data[:n]))
			return
		}
		if tag == "tdsn" {
			fmt.Printf("%schunk %s display=%q (%dB) hex=%s\n", ind, tag, parseTdsnName(c.Data), len(c.Data), hex.EncodeToString(c.Data))
			return
		}
		if tag == "Utf8" {
			fmt.Printf("%schunk %s utf8=%q\n", ind, tag, string(c.Data))
			return
		}
		snip := ""
		if len(c.Data) > 0 && len(c.Data) <= 32 {
			snip = " hex=" + hex.EncodeToString(c.Data)
		} else if len(c.Data) > 0 {
			snip = fmt.Sprintf(" (%d B) head=%s", len(c.Data), hex.EncodeToString(c.Data[:16]))
		}
		fmt.Printf("%schunk %s%s\n", ind, tag, snip)
	}
	dump(root, 0)
}
