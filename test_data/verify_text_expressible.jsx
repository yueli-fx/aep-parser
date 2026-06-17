// Differential render verify for the Expressible Selector. Reads
// text_expressible_args.json {input, done, resaved, compL, compR, pngL, pngR,
// fontSize, posX, posY}. Both comps hold "ABCDEFGH" + an Opacity-0 animator + an
// Expressible Selector; compL's Amount expression hides the LEFT half (ink ends up
// on the right), compR's hides the RIGHT half (ink on the left). Reads back each
// Amount's expression/enabled (DOM proof AE accepted it) and renders one frame of
// each so Go can assert the ink centroids are on opposite sides — proving the
// expression drives WHICH glyphs are selected.
(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/text_expressible_args.json");
    argsFile.open("r"); var raw = argsFile.read(); argsFile.close();
    var args = eval("(" + raw + ")");
    var log = []; var ok = true;
    function fail(m){ok=false;log.push("  FAIL: "+m);} function note(m){log.push("  "+m);}

    function findComp(name){
        for (var i=1;i<=app.project.numItems;i++){var it=app.project.item(i);if(it instanceof CompItem&&it.name===name)return it;}
        return null;
    }
    function setupAndRender(compName, png, rtime){
        var comp = findComp(compName);
        if(!comp){fail("comp "+compName+" not found");return;}
        var txt=null;
        for(var li=1;li<=comp.numLayers;li++){if(comp.layer(li).name==="TXT"){txt=comp.layer(li);break;}}
        if(!txt){fail("TXT missing in "+compName);return;}
        var tp=txt.property("ADBE Text Properties");
        var sels=tp.property("ADBE Text Animators").property(1).property("ADBE Text Selectors");
        var found=false;
        for(var s=1;s<=sels.numProperties;s++){
            var sel=sels.property(s);
            if(sel.matchName==="ADBE Text Expressible Selector"){
                var amt=sel.property("ADBE Text Expressible Amount");
                note(compName+" amount enabled="+amt.expressionEnabled+" expr=["+amt.expression+"]");
                if(amt.expressionEnabled!==true) fail(compName+" Amount expression not enabled");
                found=true;
            }
        }
        if(!found){fail(compName+" no Expressible Selector");return;}
        try { var st=tp.property("ADBE Text Document"); var td=st.value; if(args.fontSize) td.fontSize=args.fontSize; st.setValue(td);}catch(e){note("font err "+e);}
        try { txt.property("ADBE Transform Group").property("ADBE Position").setValue([args.posX,args.posY]); }catch(e){note("pos err "+e);}
        comp.saveFrameToPng(rtime, new File(png));
    }

    try {
        app.open(new File(args.input));
        setupAndRender(args.compL, args.pngL, 0);
        setupAndRender(args.compR, args.pngR, 1);
        app.project.save(new File(args.resaved));
        $.sleep(2000);
        if(!new File(args.pngL).exists) fail("no pngL");
        if(!new File(args.pngR).exists) fail("no pngR");
    } catch(e){ fail("EXC "+e.toString()+" line="+e.line); }

    var done=new File(args.done); done.open("w"); done.write((ok?"PASS":"FAIL")+"\n"+log.join("\n")); done.close();
    try{app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES);}catch(e){}
    try{app.quit();}catch(e){}
})();
