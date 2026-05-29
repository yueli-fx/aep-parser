package main
import("encoding/hex";"fmt";"os";"github.com/example/aep-parser/internal/rifx")
func trimNUL(s string)string{for i:=0;i<len(s);i++{if s[i]==0{return s[:i]}};return s}
func main(){
	f,_:=os.Open(os.Args[1]);defer f.Close();root,_:=rifx.Parse(f)
	names:=os.Args[2:]
	var w func(*rifx.Chunk)
	w=func(c *rifx.Chunk){for i:=0;i+1<len(c.Children);i++{ch:=c.Children[i]
		if ch.ID==rifx.IDTdmn{n:=trimNUL(string(ch.Data));for _,want:=range names{if n==want{
			fmt.Printf("\n=== %s ===\n",n)
			for _,cc:=range c.Children[i+1].Children{if cc.IsList(){for _,x:=range cc.Children{
				if x.ID==rifx.IDLhd3{fmt.Printf("lhd3 bpk@0x10=%d:\n%s",func()uint32{var v uint32;for b:=0;b<4;b++{v=v<<8|uint32(x.Data[0x10+b])};return v}(),hex.Dump(x.Data[:24]))}
				if x.ID==rifx.IDLdat{fmt.Printf("ldat %dB:\n%s",len(x.Data),hex.Dump(x.Data))}}}}
		}}}
		if ch.IsList(){w(ch)}}
		if n:=len(c.Children);n>0&&c.Children[n-1].IsList(){w(c.Children[n-1])}}
	w(root)
}
