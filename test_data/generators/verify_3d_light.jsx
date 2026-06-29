// Ship gate for from-scratch 3D lighting. Reads 3d_light_args.json {input, done,
// resaved, png}. The scene (a 3D gray PANEL lit by a from-scratch POINT light
// placed up-left/in-front) is Go-built; this JSX reads back the light's
// type=POINT + the panel's threeDLayer flag, resaves, and renders frame 0 so Go
// can assert on pixels that the panel carries a brightness falloff (bright near
// the light, dim far from it) — the visible signature of lighting on a 3D layer.
(function () {
    var a = new File("e:/projects/tools/aep-parser/test_data/generated/args/3d_light_args.json");
    a.open("r"); var raw = a.read(); a.close();
    var args = eval("(" + raw + ")");
    var log = []; var ok = true;
    function fail(m){ok=false;log.push("  FAIL: "+m);} function note(m){log.push("  "+m);}
    function lt(v){
        if (v === LightType.PARALLEL) return "PARALLEL";
        if (v === LightType.SPOT) return "SPOT";
        if (v === LightType.POINT) return "POINT";
        if (v === LightType.AMBIENT) return "AMBIENT";
        return "?"+v;
    }
    try {
        app.open(new File(args.input));
        var comp = null;
        for (var i=1;i<=app.project.numItems;i++){var it=app.project.item(i);if(it instanceof CompItem&&it.name==="LIT3D"){comp=it;break;}}
        if(!comp){fail("comp LIT3D not found");}
        else {
            for (var li=1; li<=comp.numLayers; li++){
                var L=comp.layer(li);
                if (L instanceof LightLayer){
                    note("light type="+lt(L.lightType));
                    if (L.lightType !== LightType.POINT) fail("light not POINT (got "+lt(L.lightType)+")");
                } else if (L.name==="PANEL"){
                    note("PANEL 3D="+L.threeDLayer);
                    if (!L.threeDLayer) fail("PANEL not 3D");
                }
            }
            app.project.save(new File(args.resaved));
            app.project.bitsPerChannel=8;
            comp.saveFrameToPng(0,new File(args.png));
            $.sleep(2000);
            if(!new File(args.png).exists) fail("png not written");
        }
    } catch(e){ fail("EXC "+e.toString()+" line="+e.line); }
    var d=new File(args.done); d.open("w"); d.write((ok?"PASS":"FAIL")+"\n"+log.join("\n")); d.close();
    try{app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES);}catch(e){}
    try{app.quit();}catch(e){}
})();
