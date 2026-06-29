// RE/extract fixture for the Expressible Selector. Builds a text layer with an
// Opacity-0 animator + an "ADBE Text Expressible Selector", logs AE's default
// Expressible Amount expression/enabled state, and saves re_text_expressible.aep
// for tmp_debug/extract_text_animator to pull the selector body template.
(function () {
    var out = new File("e:/projects/tools/aep-parser/test_data/fixtures/re_text_expressible.aep");
    var done = new File("e:/projects/tools/aep-parser/test_data/re_text_expressible.done");
    var log = [];
    try {
        app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES);
        app.newProject();
        var comp = app.project.items.addComp("EXPRSEL", 1280, 720, 1, 5, 24);
        var t = comp.layers.addText("ABCDEF");
        var anim = t.property("ADBE Text Properties").property("ADBE Text Animators").addProperty("ADBE Text Animator");
        anim.property("ADBE Text Animator Properties").addProperty("ADBE Text Opacity").setValue(0);
        var sels = anim.property("ADBE Text Selectors");
        var es = sels.addProperty("ADBE Text Expressible Selector");
        var amt = es.property("ADBE Text Expressible Amount");
        log.push("amount matchName=" + amt.matchName);
        log.push("amount canSetExpression=" + amt.canSetExpression);
        log.push("amount expressionEnabled=" + amt.expressionEnabled);
        log.push("amount expression=[" + amt.expression + "]");
        // Materialize the Amount param: setting an expression forces AE to persist
        // the param's tdbs (otherwise it is fully elided), giving the extracted
        // template a slot AddTextExpressibleSelector can overwrite via SetExpression.
        amt.expression = "selectorValue";
        log.push("after set: enabled=" + amt.expressionEnabled + " expr=[" + amt.expression + "]");
        app.project.save(out);
        log.push("saved " + out.fsName);
    } catch (e) { log.push("EXC " + e.toString() + " line=" + e.line); }
    try { done.open("w"); done.write(log.join("\n")); done.close(); } catch (e) {}
    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
