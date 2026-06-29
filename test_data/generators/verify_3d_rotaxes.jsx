// Ship gate for 3D Rotate X / Orientation / Rotate Z on a from-scratch 3D layer
// (roadmap priority 2 closeout — same zero-new-write-code path as Rotate Y).
// Reads 3d_rotaxes_args.json {input, done, resaved, png, comp, prop, idx, expect,
// time}. The scene (one 3D box rotated about ONE axis + a camera) is Go-built;
// this JSX reads back the rotation channel, resaves, and renders the frame at a
// UNIQUE per-axis time (disk-cache key isolation — comp.id is always 1 for a
// from-scratch single-comp project, so distinct times avoid cross-process stale
// frames) so Go can assert on pixels that the box is foreshortened/rotated.
(function () {
    var a = new File("e:/projects/tools/aep-parser/test_data/generated/args/3d_rotaxes_args.json");
    a.open("r"); var raw = a.read(); a.close();
    var args = eval("(" + raw + ")");
    var log = []; var ok = true;
    function fail(m){ok=false;log.push("  FAIL: "+m);} function note(m){log.push("  "+m);}
    try {
        app.open(new File(args.input));
        var comp = null;
        for (var i=1;i<=app.project.numItems;i++){var it=app.project.item(i);if(it instanceof CompItem&&it.name===args.comp){comp=it;break;}}
        if(!comp){fail("comp "+args.comp+" not found");}
        else {
            for (var li=1; li<=comp.numLayers; li++){
                var L=comp.layer(li);
                if (L.name==="BOX"){
                    var pv=L.property("ADBE Transform Group").property(args.prop);
                    var val=pv?pv.value:null;
                    var got=(val instanceof Array)?val[args.idx]:val;
                    note("BOX 3D="+L.threeDLayer+" "+args.prop+"["+args.idx+"]="+got);
                    if (L.threeDLayer!==true) fail("BOX not 3D");
                    if (got===null || Math.abs(got-args.expect)>0.5) fail(args.prop+" != "+args.expect+" (got "+got+")");
                }
            }
            app.project.save(new File(args.resaved));
            app.project.bitsPerChannel=8;
            comp.saveFrameToPng(args.time,new File(args.png));
            $.sleep(1500);
            if(!new File(args.png).exists) fail("png not written");
        }
    } catch(e){ fail("EXC "+e.toString()+" line="+e.line); }
    var d=new File(args.done); d.open("w"); d.write((ok?"PASS":"FAIL")+"\n"+log.join("\n")); d.close();
    try{app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES);}catch(e){}
    try{app.quit();}catch(e){}
})();
