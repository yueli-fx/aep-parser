// verify_v2_2_pathkf.jsx — AE-side ship gate for an ANIMATED Path ShapeLayer.
// Reads args JSON, opens the builder .aep, asserts the Path stream ("ADBE
// Vector Shape") is animated with >=2 keyframes (layer not dropped, animation
// retained), re-saves so Go can re-decode AE's canonical om-s, writes PASS/FAIL.
#target aftereffects

(function () {
	function readArgs() {
		var argsFile = new File($.fileName).parent.fsName + "/v2_2_pathkf_args.json";
		var f = new File(argsFile);
		f.open("r");
		var raw = f.read();
		f.close();
		return eval("(" + raw + ")");
	}

	function fail(doneFile, msg) {
		var f = new File(doneFile);
		f.open("w");
		f.write("FAIL\n" + msg);
		f.close();
	}

	function pass(doneFile) {
		var f = new File(doneFile);
		f.open("w");
		f.write("PASS\n");
		f.close();
	}

	// findByMatch: depth-first search for a property by matchName. The Path
	// stream ("ADBE Vector Shape") sits 3 levels below Root Vectors Group
	// (Vector Group → Vectors Group → Vector Shape - Group → Vector Shape), so
	// a flat property(1).property(name) lookup misses it.
	function findByMatch(grp, mn) {
		if (!grp || !grp.numProperties) return null;
		for (var i = 1; i <= grp.numProperties; i++) {
			var p = grp.property(i);
			if (p.matchName === mn) return p;
			if (p.numProperties) {
				var r = findByMatch(p, mn);
				if (r) return r;
			}
		}
		return null;
	}

	var args = null;
	try {
		args = readArgs();
		var proj = app.open(new File(args.input));
		if (!proj || proj.numItems < 1) throw "no items in project";
		var comp = null;
		for (var i = 1; i <= proj.numItems; i++) {
			if (proj.item(i) instanceof CompItem) { comp = proj.item(i); break; }
		}
		if (!comp) throw "no comp";
		if (comp.numLayers < 1) throw "comp has 0 layers (shape layer dropped)";
		var layer = comp.layer(1);
		var contents = layer.property("ADBE Root Vectors Group");
		if (!contents) throw "no contents";
		var pathProp = findByMatch(contents, "ADBE Vector Shape");
		if (!pathProp) throw "no path shape stream";
		// >=6 proves the 6-keyframe build survived (paging works; a corrupt
		// lhd3 capacity would make AE 2025 reject the whole file before this).
		if (!pathProp.numKeys || pathProp.numKeys < 6)
			throw "path animation truncated (numKeys=" + (pathProp.numKeys || 0) + ", want >=6)";
		// The t=2 keyframe (key index 3, time-sorted 0,1,2,3,4,4.5) was built with
		// temporal ease → AE must read it as BEZIER (not LINEAR). Proves
		// encodePathTimeTable's ease write was ingested, not silently dropped.
		var k = 3;
		if (pathProp.keyInInterpolationType(k) !== KeyframeInterpolationType.BEZIER ||
			pathProp.keyOutInterpolationType(k) !== KeyframeInterpolationType.BEZIER)
			throw "eased keyframe " + k + " not BEZIER (in=" + pathProp.keyInInterpolationType(k) +
				" out=" + pathProp.keyOutInterpolationType(k) + ", BEZIER=" + KeyframeInterpolationType.BEZIER + ")";
		proj.save(new File(args.resaved));
		pass(args.done);
	} catch (e) {
		fail(args ? args.done : (new File($.fileName).parent.fsName + "/v2_2_pathkf.done"), String(e));
	}
})();
