// Generic AE render oracle entrypoint.
//
// Reads a JSON RenderRequest from the AEORACLE_REQUEST environment variable.
// If the variable is absent, falls back to aeoracle_request.json beside this JSX.
// Intended to run through scripts/ae-worker/ae_run.ps1, which watches request.done_path.
(function () {
    function installJSON() {
        if (typeof JSON === "undefined") JSON = {};
        if (typeof JSON.parse !== "function") {
            JSON.parse = function (text) {
                return eval("(" + text + ")");
            };
        }
        if (typeof JSON.stringify === "function") return;

        function repeat(text, count) {
            var out = "";
            for (var i = 0; i < count; i++) out += text;
            return out;
        }

        function quote(text) {
            return "\"" + String(text)
                .replace(/\\/g, "\\\\")
                .replace(/"/g, "\\\"")
                .replace(/\r/g, "\\r")
                .replace(/\n/g, "\\n")
                .replace(/\t/g, "\\t") + "\"";
        }

        function encode(value, indent, depth) {
            if (value === null) return "null";
            var kind = typeof value;
            if (kind === "string") return quote(value);
            if (kind === "number") return isFinite(value) ? String(value) : "null";
            if (kind === "boolean") return value ? "true" : "false";
            if (kind === "undefined" || kind === "function") return undefined;

            var gap = indent ? repeat(indent, depth) : "";
            var nextGap = indent ? repeat(indent, depth + 1) : "";
            var parts = [];
            var i;
            if (value instanceof Array) {
                for (i = 0; i < value.length; i++) {
                    var item = encode(value[i], indent, depth + 1);
                    parts.push(item === undefined ? "null" : item);
                }
                if (!indent) return "[" + parts.join(",") + "]";
                return "[\n" + nextGap + parts.join(",\n" + nextGap) + "\n" + gap + "]";
            }

            for (var key in value) {
                if (!value.hasOwnProperty(key)) continue;
                var encoded = encode(value[key], indent, depth + 1);
                if (encoded === undefined) continue;
                parts.push(quote(key) + (indent ? ": " : ":") + encoded);
            }
            if (!indent) return "{" + parts.join(",") + "}";
            return "{\n" + nextGap + parts.join(",\n" + nextGap) + "\n" + gap + "}";
        }

        JSON.stringify = function (value, replacer, space) {
            var indent = "";
            if (typeof space === "number") indent = repeat(" ", Math.min(space, 10));
            else if (typeof space === "string") indent = space.substring(0, 10);
            return encode(value, indent, 0);
        };
    }

    installJSON();

    var requestPath = $.getenv("AEORACLE_REQUEST");
    if (!requestPath) {
        var self = new File($.fileName);
        requestPath = self.parent.fsName + "/aeoracle_request.json";
    }
    var baseDir = $.getenv("AEORACLE_CWD") || (new File(requestPath)).parent.fsName;

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

    function isAbsolutePath(path) {
        path = slash(path);
        return /^[A-Za-z]:\//.test(path) || path.indexOf("//") === 0 || path.charAt(0) === "/";
    }

    function resolvePath(path) {
        if (!path) return path;
        path = String(path);
        if (isAbsolutePath(path)) return path;
        return joinPath(baseDir, path);
    }

    function waitForNonEmptyFile(path, timeoutMs) {
        var deadline = (new Date()).getTime() + timeoutMs;
        var f = new File(path);
        while ((new Date()).getTime() < deadline) {
            if (f.exists && f.length > 0) return true;
            $.sleep(250);
            f = new File(path);
        }
        return f.exists && f.length > 0;
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
        req.aep_path = resolvePath(req.aep_path);
        req.output_dir = resolvePath(req.output_dir);
        req.done_path = resolvePath(req.done_path);
        req.metadata_path = resolvePath(req.metadata_path);
        var frameTimeoutMs = Number(req.frame_timeout_ms || 120000);
        if (!isFinite(frameTimeoutMs) || frameTimeoutMs <= 0) frameTimeoutMs = 120000;
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
            if (!waitForNonEmptyFile(outPath, frameTimeoutMs)) {
                throw new Error("rendered frame missing or empty: " + outPath);
            }
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
