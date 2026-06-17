// Ship gate for per-layer 3D Z (parallax). Reads 3d_parallax_args.json
// {input, done, resaved, png}. The whole scene (two same-size 3D boxes at
// different Z + a camera) is Go-built; this JSX only reads back the boxes' 3D
// state + Z, resaves, and renders frame 0 so Go can assert on pixels that the
// near box renders larger than the far box (depth from per-layer Z).
(function () {
    var a = new File("e:/projects/tools/aep-parser/test_data/3d_parallax_args.json");
    a.open("r"); var raw = a.read(); a.close();
    var args = eval("(" + raw + ")");
    var log = []; var ok = true;
    function fail(m){ok=false;log.push("  FAIL: "+m);} function note(m){log.push("  "+m);}
    try {
        app.open(new File(args.input));
        var comp = null;
        for (var i=1;i<=app.project.numItems;i++){var it=app.project.item(i);if(it instanceof CompItem&&it.name==="CAM3DP"){comp=it;break;}}
        if(!comp){fail("comp CAM3DP not found");}
        else {
            for (var li=1; li<=comp.numLayers; li++){
                var L=comp.layer(li);
                if (L.name==="NEAR"||L.name==="FAR"){
                    var pos=L.property("ADBE Transform Group").property("ADBE Position");
                    note(L.name+" 3D="+L.threeDLayer+" pos="+pos.value.toString());
                    if (L.threeDLayer!==true) fail(L.name+" not 3D");
                    if (pos.value.length!==3) fail(L.name+" Position not 3D");
                }
            }
            app.project.save(new File(args.resaved));
            app.project.bitsPerChannel=8;
            comp.saveFrameToPng(0,new File(args.png));
            $.sleep(1500);
            if(!new File(args.png).exists) fail("png not written");
        }
    } catch(e){ fail("EXC "+e.toString()+" line="+e.line); }
    var d=new File(args.done); d.open("w"); d.write((ok?"PASS":"FAIL")+"\n"+log.join("\n")); d.close();
    try{app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES);}catch(e){}
    try{app.quit();}catch(e){}
})();
