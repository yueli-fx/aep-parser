package main
import ("fmt";"os";"strings";"github.com/example/aep-parser/internal/rifx")
func main(){
 f,_:=os.Open(os.Args[1]); defer f.Close()
 root,_:=rifx.Parse(f)
 var walk func(c *rifx.Chunk, d int)
 walk=func(c *rifx.Chunk, d int){
  for _,ch:=range c.Children {
   if ch.ID==rifx.IDTdmn {
    nm:=strings.TrimRight(string(ch.Data),"\x00")
    if nm!="" && nm!="ADBE Group End" { fmt.Printf("%s%s\n", strings.Repeat("  ",d), nm) }
   }
   if ch.IsList(){ walk(ch, d+1) }
  }
 }
 walk(root,0)
}
