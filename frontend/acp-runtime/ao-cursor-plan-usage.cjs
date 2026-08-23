"use strict";

const fs = require("node:fs");
const Module = require("node:module");
const path = require("node:path");

function patchAboutSource(source) {
	const importNeedle = ',u=l("./src/utils/terminal-environment.ts"),d=function';
	const importReplacement = ',u=l("./src/utils/terminal-environment.ts"),usageModule=l.e(6260).then(()=>l("./src/usage/usage-data.ts")),d=function';
	if (!source.includes(importNeedle)) throw new Error("unsupported Cursor about module imports");
	source = source.replace(importNeedle, importReplacement);

	const returnNeedle = "return{cliVersion:f,model:b,subscriptionTier:A,osPlatform:p,osArch:h,userEmail:y,terminalProgram:q,shell:I,lastRequestId:R}";
	const returnReplacement = 'const usage=yield usageModule,usageResult=yield usage.d({client:l,locale:"en-US"});return{cliVersion:f,usage:usageResult}';
	if (!source.includes(returnNeedle)) throw new Error("unsupported Cursor about module result");
	return source.replace(returnNeedle, returnReplacement);
}

function run(cursorDir) {
	if (!cursorDir) throw new Error("Cursor runtime directory is required");
	const entry = path.join(cursorDir, "index.js");
	const aboutChunk = path.join(cursorDir, "5105.index.js");
	const originalLoader = Module._extensions[".js"];
	Module._extensions[".js"] = function cursorUsageLoader(module, filename) {
		if (filename === aboutChunk) {
			module._compile(patchAboutSource(fs.readFileSync(filename, "utf8")), filename);
			return;
		}
		originalLoader(module, filename);
	};
	process.argv = [process.execPath, entry, "about", "--format", "json"];
	require(entry);
}

module.exports = { patchAboutSource };

if (require.main === module) {
	run(process.argv[2]);
}
