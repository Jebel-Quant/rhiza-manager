const esbuild = require("esbuild");

esbuild.build({
  entryPoints: ["src/extension.ts"],
  bundle: true,
  platform: "node",
  outfile: "dist/extension.js",
  external: ["vscode"],
  sourcemap: true
}).catch(() => process.exit(1));

