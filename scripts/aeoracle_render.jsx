// Generic AE render oracle entrypoint.
//
// Reads a JSON RenderRequest from the AEORACLE_REQUEST environment variable.
// If the variable is absent, falls back to aeoracle_request.json beside this JSX.
// Intended to run through scripts/ae_run.ps1, which watches request.done_path.
(function () {
    var requestPath = $.getenv("AEORACLE_REQUEST");
    if (!requestPath) {
        var self = new File($.fileName);
        requestPath = self.parent.fsName + "/aeoracle_request.json";
    }

    function readText(path) {
        var f = new File(path);
        f.encoding = "UTF-8";
        if (!f.open("r")) throw new Error("cannot open request " + path);
        var text = f.read();
        f.close();
        return text;
    }

    function writeText(path, text) {
        var f = new File(path);
        f.encoding = "UTF-8";
        var parent = f.parent;
        if (parent && !parent.exists) parent.create();
        if (!f.open("w")) throw new Error("cannot write " + path);
        f.write(text);
        f.close();
    }

    function slash(path) {
        return String(path || "").replace(/\\/g, "/");
    }

    function joinPath(dir, name) {
        dir = slash(dir);
        if (dir.charAt(dir.length - 1) !== "/") dir += "/";
        return dir + name;
    }

    function findComp(name) {
        var first = null;
        for (var i = 1; i <= app.project.numItems; i++) {
            var it = app.project.item(i);
            if (!(it instanceof CompItem)) continue;
            if (!first) first = it;
            if (name && it.name === name) return it;
        }
        return name ? null : first;
    }

    var req = null;
    var metadata = {
        schema_version: 1,
        status: "started",
        frames: [],
        warnings: []
    };

    try {
        req = JSON.parse(readText(requestPath));
        metadata.aep_path = req.aep_path;
        metadata.comp_name = req.comp_name || "";
        metadata.output_dir = req.output_dir;
        metadata.ae_version = app.version;
        metadata.os = $.os;

        var outFolder = new Folder(req.output_dir);
        if (!outFolder.exists) outFolder.create();

        app.open(new File(req.aep_path));
        app.project.bitsPerChannel = 8;
        var comp = findComp(req.comp_name || "");
        if (!comp) throw new Error("comp not found: " + req.comp_name);
        metadata.comp_name = comp.name;

        for (var i = 0; i < req.frames.length; i++) {
            var frame = req.frames[i];
            var outPath = joinPath(req.output_dir, frame.tag + ".png");
            try {
                app.purge(PurgeTarget.ALL_CACHES);
            } catch (purgeErr) {}
            comp.saveFrameToPng(frame.seconds, new File(outPath));
            $.sleep(500);
            metadata.frames.push({
                frame: frame.frame,
                seconds: frame.seconds,
                tag: frame.tag,
                reason: frame.reason,
                output_path: outPath,
                status: "rendered"
            });
        }
        metadata.status = "ok";
    } catch (e) {
        metadata.status = "error";
        metadata.warnings.push(String(e) + (e.line ? (" line=" + e.line) : ""));
    }

    try {
        var metaPath = req && req.metadata_path ? req.metadata_path : joinPath((req && req.output_dir) || (new File(requestPath)).parent.fsName, "metadata.json");
        writeText(metaPath, JSON.stringify(metadata, null, 2));
    } catch (metaErr) {}

    try {
        var donePath = req && req.done_path ? req.done_path : joinPath((new File(requestPath)).parent.fsName, "aeoracle_render.done");
        writeText(donePath, metadata.status);
    } catch (doneErr) {}

    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (closeErr) {}
    try { app.quit(); } catch (quitErr) {}
})();
