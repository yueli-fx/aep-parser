// Ship gate for from-scratch 3D shadows. Reads 3d_shadow_args.json {input, done,
// resaved, png}. The scene (a 3D WALL catcher + a 3D CASTER with synthesized
// Material Casts Shadows=On + a POINT light with Casts Shadows=On) is Go-built;
// this JSX reads back the caster's material Casts Shadows + the light's Casts
// Shadows, resaves, and renders frame 0 so Go can assert on pixels that a shadow
// darkens the wall.
(function () {
    var a = new File("e:/projects/tools/aep-parser/test_data/3d_shadow_args.json");
    a.open("r"); var raw = a.read(); a.close();
    var args = eval("(" + raw + ")");
    var log = []; var ok = true;
    function fail(m){ok=false;log.push("  FAIL: "+m);} function note(m){log.push("  "+m);}
    try {
        app.open(new File(args.input));
        var comp = null;
        for (var i=1;i<=app.project.numItems;i++){var it=app.project.item(i);if(it instanceof CompItem&&it.name==="SHADOW3D"){comp=it;break;}}
        if(!comp){fail("comp SHADOW3D not found");}
        else {
            for (var li=1; li<=comp.numLayers; li++){
                var L=comp.layer(li);
                if (L instanceof LightLayer){
                    var lo=L.property("ADBE Light Options Group");
                    var lcs=lo?lo.property("ADBE Casts Shadows"):null;
                    note("light castsShadows="+(lcs?lcs.value:"nil"));
                    if(!lcs||lcs.value!==1) fail("light not casting shadows");
                } else if (L.name==="CASTER"){
                    var mo=L.property("ADBE Material Options Group");
                    var mcs=mo?mo.property("ADBE Casts Shadows"):null;
                    note("CASTER 3D="+L.threeDLayer+" materialCastsShadows="+(mcs?mcs.value:"nil"));
                    if(!mcs||mcs.value!==1) fail("caster material not casting shadows");
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
