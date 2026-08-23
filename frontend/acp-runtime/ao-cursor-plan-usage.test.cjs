"use strict";

const assert = require("node:assert/strict");
const test = require("node:test");
const { patchAboutSource } = require("./ao-cursor-plan-usage.cjs");

test("patchAboutSource adds only the sanitized usage model", () => {
	const source = 'x,u=l("./src/utils/terminal-environment.ts"),d=function y return{cliVersion:f,model:b,subscriptionTier:A,osPlatform:p,osArch:h,userEmail:y,terminalProgram:q,shell:I,lastRequestId:R} z';
	const patched = patchAboutSource(source);
	assert.match(patched, /usageModule=l\.e\(6260\)/);
	assert.match(patched, /return\{cliVersion:f,usage:usageResult\}/);
	assert.doesNotMatch(patched, /userEmail/);
});

test("patchAboutSource rejects an unknown Cursor bundle shape", () => {
	assert.throws(() => patchAboutSource("different build"), /unsupported Cursor/);
});
