// AE 24+ variable font axes RE 工程 (round 3).
//
// 已确认：
//   * axes 存在 btdk `/0/1/0[i]/0/0/4` 数组 (16.16 fixed-point；
//     Bahnschrift = [wght=400, wdth=100] = [2.62144e7, 6.5536e6])
//   * `td.fontVariation = X` 是 no-op (TextDocument 没这个 key)
//   * 正确 API 入口: app.fonts.getFontsByFamilyNameAndStyleName + td.fontObject
//
// Round 3：用真正的 Font API 构造 variable-font instance，对比 btdk 看哪里写 axes:
//   vf_baseline    — TimesNewRomanPSMT (非 variable 对照)
//   vf_default     — Bahnschrift Regular (默认 axes)
//   vf_bold        — Bahnschrift Bold (named instance, wght=700)
//   vf_light_cond  — Bahnschrift Condensed Light (named instance, wght=300, wdth=75)

(function () {
    var outFile = new File("e:/projects/tools/aep-parser/test_data/fixtures/re_varfont_ae24.aep");
    var log = [];
    function step(name, fn) {
        try { fn(); log.push("OK  " + name); }
        catch (e) { log.push("ERR " + name + " -> " + e.toString()); }
    }

    // Start fresh: close any prior project so accumulated `RE_VARFONT`
    // comps from earlier rounds don't pile up. AE accepts duplicate comp
    // names, so without this each rerun adds new instances.
    step("fresh_project", function () {
        app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES);
        app.newProject();
        app.project.save(outFile);
    });

    var comp;
    step("add_comp", function () {
        comp = app.project.items.addComp("RE_VARFONT", 1920, 1080, 1, 5, 30);
    });

    // List all Bahnschrift styles so we know which named-instance strings work.
    // (Kept for diagnostics — the round-1 result said this filter returns empty,
    // so the family-name match against Font objects may use a different attribute;
    // resolution by family+style still works downstream.)
    step("list_bahnschrift_styles", function () {
        var fonts = app.fonts.allFonts;
        var bs = [];
        for (var i = 0; i < fonts.length; i++) {
            var f = fonts[i];
            if (f.familyName === "Bahnschrift") {
                bs.push(f.styleName + " | ps=" + f.postScriptName);
            }
        }
        log.push("  Bahnschrift styles: " + bs.join(" / "));
    });

    function addLayer(name, fontPS) {
        var tl = comp.layers.addText("Vf");
        tl.name = name;
        var td = tl.sourceText.value;
        try { td.font = fontPS; }
        catch (e) { log.push("  " + name + ".font='" + fontPS + "' err: " + e); }
        tl.sourceText.setValue(td);
        return tl;
    }

    function addLayerByFamily(name, familyName, styleName) {
        var tl = comp.layers.addText("Vf");
        tl.name = name;
        var fonts = app.fonts.getFontsByFamilyNameAndStyleName(familyName, styleName);
        if (!fonts || fonts.length === 0) {
            log.push("  " + name + ": no font found for '" + familyName + "' / '" + styleName + "'");
            return tl;
        }
        var fontPS = fonts[0].postScriptName;
        log.push("  " + name + " resolved PS=" + fontPS);
        var td = tl.sourceText.value;
        try { td.font = fontPS; }
        catch (e) { log.push("  " + name + ".font='" + fontPS + "' err: " + e); }
        tl.sourceText.setValue(td);
        return tl;
    }

    step("vf_baseline", function () { addLayer("vf_baseline", "TimesNewRomanPSMT"); });
    step("vf_default",  function () { addLayerByFamily("vf_default", "Bahnschrift", "Regular"); });
    step("vf_bold",     function () { addLayerByFamily("vf_bold", "Bahnschrift", "Bold"); });
    step("vf_light_cond", function () { addLayerByFamily("vf_light_cond", "Bahnschrift", "Condensed Light"); });

    step("save", function () { app.project.save(); });

    step("write_done_marker", function () {
        var marker = new File("e:/projects/tools/aep-parser/test_data/re_varfont_ae24.done");
        marker.open("w");
        marker.write(log.join("\n"));
        marker.close();
    });
})();
