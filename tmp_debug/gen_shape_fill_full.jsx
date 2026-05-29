// V2.2.1 子项⑨ richer fill body template source: single shape layer, Fill with
// Opacity set static non-default (60) so AE emits an Opacity cdat slot (default
// 100 is elided). Extract via extract_shape_bodies → v2_2_shape_fill_body.bin.
(function(){var o=new File("e:/projects/tools/aep-parser/test_data/v2_2_shape_fill_full.aep"),d=new File("e:/projects/tools/aep-parser/test_data/v2_2_shape_fill_full.done"),L=[],ok=false;
try{app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES);app.newProject();var c=app.project.items.addComp("FillFull",1920,1080,1,5,30);var s=c.layers.addShape();s.name="Nested";var sc=s.property("ADBE Root Vectors Group").addProperty("ADBE Vector Group").property("ADBE Vectors Group");sc.addProperty("ADBE Vector Shape - Rect").property("ADBE Vector Rect Size").setValue([200,100]);var f=sc.addProperty("ADBE Vector Graphic - Fill");f.property("ADBE Vector Fill Color").setValue([0.5,0.5,0.5,1]);f.property("ADBE Vector Fill Opacity").setValue(60);app.project.save(o);L.push("saved");ok=true;}catch(e){L.push("ERROR: "+e);}
try{d.open("w");d.write((ok?"PASS\n":"FAIL\n")+L.join("\n"));d.close();}catch(e2){}
try{if(app.project)app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES);}catch(e3){}try{app.quit();}catch(e4){}})();
