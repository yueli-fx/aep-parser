// Ship gate for 3D Rotate Y (perspective tumble). Reads 3d_roty_args.json
// {input, done, resaved, png}. The scene (one 3D box rotated 50° about Y + a
// camera) is Go-built; this JSX reads back the box's RotateY, resaves, and
// renders frame 0 so Go can assert on pixels that the box renders as a
// foreshortened trapezoid (near edge taller than far edge) — true perspective.
(function () {
    var a = new File("e:/projects/tools/aep-parser/test_data/3d_roty_args.json");
    a.open("r"); var raw = a.read(); a.close();
    var args = eval("(" + raw + ")");
    var log = []; var ok = true;
    function fail(m){ok=false;log.push("  FAIL: "+m);} function note(m){log.push("  "+m);}
    try {
        app.open(new File(args.input));
        var comp = null;
        for (var i=1;i<=app.project.numItems;i++){var it=app.project.item(i);if(it instanceof CompItem&&it.name==="ROTY"){comp=it;break;}}
        if(!comp){fail("comp ROTY not found");}
        else {
            for (var li=1; li<=comp.numLayers; li++){
                var L=comp.layer(li);
                if (L.name==="BOX"){
                    var ry=L.property("ADBE Transform Group").property("ADBE Rotate Y");
                    note("BOX 3D="+L.threeDLayer+" RotateY="+(ry?ry.value:"nil"));
                    if (L.threeDLayer!==true) fail("BOX not 3D");
                    if (!ry || Math.abs(ry.value-50)>0.5) fail("RotateY != 50 (got "+(ry?ry.value:"nil")+")");
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
