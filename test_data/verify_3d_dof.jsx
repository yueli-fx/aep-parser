// Ship gate for camera Depth of Field. Reads 3d_dof_args.json {input, done,
// resaved, png}. The scene (a near SHARP box at the focus distance + a far BLUR
// box, under a camera with DoF on) is Go-built; this JSX reads back the camera's
// DoF=on + focus distance, resaves, and renders frame 0 so Go can assert on
// pixels that the out-of-focus far box is blurred while the focused near box is
// crisp.
(function () {
    var a = new File("e:/projects/tools/aep-parser/test_data/3d_dof_args.json");
    a.open("r"); var raw = a.read(); a.close();
    var args = eval("(" + raw + ")");
    var log = []; var ok = true;
    function fail(m){ok=false;log.push("  FAIL: "+m);} function note(m){log.push("  "+m);}
    try {
        app.open(new File(args.input));
        var comp = null;
        for (var i=1;i<=app.project.numItems;i++){var it=app.project.item(i);if(it instanceof CompItem&&it.name==="DOF"){comp=it;break;}}
        if(!comp){fail("comp DOF not found");}
        else {
            for (var li=1; li<=comp.numLayers; li++){
                var L=comp.layer(li);
                if (L instanceof CameraLayer){
                    var o=L.property("ADBE Camera Options Group");
                    var dof=o.property("ADBE Camera Depth of Field");
                    var fd=o.property("ADBE Camera Focus Distance");
                    note("camera DoF="+(dof?dof.value:"nil")+" focusDistance="+(fd?fd.value:"nil"));
                    if (!dof || dof.value!==1) fail("camera DoF not enabled");
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
