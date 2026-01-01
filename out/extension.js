"use strict";
var __create = Object.create;
var __defProp = Object.defineProperty;
var __getOwnPropDesc = Object.getOwnPropertyDescriptor;
var __getOwnPropNames = Object.getOwnPropertyNames;
var __getProtoOf = Object.getPrototypeOf;
var __hasOwnProp = Object.prototype.hasOwnProperty;
var __export = (target, all) => {
  for (var name in all)
    __defProp(target, name, { get: all[name], enumerable: true });
};
var __copyProps = (to, from, except, desc) => {
  if (from && typeof from === "object" || typeof from === "function") {
    for (let key of __getOwnPropNames(from))
      if (!__hasOwnProp.call(to, key) && key !== except)
        __defProp(to, key, { get: () => from[key], enumerable: !(desc = __getOwnPropDesc(from, key)) || desc.enumerable });
  }
  return to;
};
var __toESM = (mod, isNodeMode, target) => (target = mod != null ? __create(__getProtoOf(mod)) : {}, __copyProps(
  // If the importer is in node compatibility mode or this is not an ESM
  // file that has been converted to a CommonJS file using a Babel-
  // compatible transform (i.e. "__esModule" has not been set), then set
  // "default" to the CommonJS "module.exports" for node compatibility.
  isNodeMode || !mod || !mod.__esModule ? __defProp(target, "default", { value: mod, enumerable: true }) : target,
  mod
));
var __toCommonJS = (mod) => __copyProps(__defProp({}, "__esModule", { value: true }), mod);

// src/extension.ts
var extension_exports = {};
__export(extension_exports, {
  activate: () => activate,
  deactivate: () => deactivate
});
module.exports = __toCommonJS(extension_exports);
var vscode = __toESM(require("vscode"));
var import_child_process = require("child_process");
var fs = __toESM(require("fs"));
var path = __toESM(require("path"));
function runGitCommand(repoPath, command) {
  return new Promise((resolve, reject) => {
    (0, import_child_process.exec)(command, { cwd: repoPath }, (err, stdout, stderr) => {
      if (err) {
        reject(stderr || err.message);
      } else {
        resolve(stdout.trim());
      }
    });
  });
}
async function getRepoStatus(repoPath) {
  try {
    const branch = await runGitCommand(repoPath, "git branch --show-current") || "detached";
    const dirtyOutput = await runGitCommand(repoPath, "git status --porcelain");
    const dirty = dirtyOutput.length > 0;
    let ahead = 0, behind = 0;
    try {
      const revList = await runGitCommand(
        repoPath,
        "git rev-list --left-right --count HEAD...@{upstream}"
      );
      [ahead, behind] = revList.split("	").map(Number);
    } catch {
    }
    return { branch, dirty, ahead, behind };
  } catch {
    return { branch: "unknown", dirty: false, ahead: 0, behind: 0 };
  }
}
var RepoItem = class extends vscode.TreeItem {
  constructor(repoPath, label, status) {
    super(label, vscode.TreeItemCollapsibleState.None);
    this.repoPath = repoPath;
    this.label = label;
    this.status = status;
    this.tooltip = repoPath;
    this.description = status;
    this.contextValue = "repo";
  }
};
var RepoProvider = class {
  _onDidChangeTreeData = new vscode.EventEmitter();
  onDidChangeTreeData = this._onDidChangeTreeData.event;
  refresh() {
    this._onDidChangeTreeData.fire(void 0);
  }
  getTreeItem(item) {
    return item;
  }
  async getChildren() {
    const folders = vscode.workspace.workspaceFolders;
    if (!folders) return [];
    const repos = [];
    for (const folder of folders) {
      const root = folder.uri.fsPath;
      for (const entry of fs.readdirSync(root)) {
        const fullPath = path.join(root, entry);
        if (fs.statSync(fullPath).isDirectory() && fs.existsSync(path.join(fullPath, ".git"))) {
          const status = await getRepoStatus(fullPath);
          const desc = `${status.branch} \xB7 ${status.dirty ? "dirty" : "clean"} \xB7 \u2191${status.ahead} \u2193${status.behind}`;
          repos.push(new RepoItem(fullPath, entry, desc));
        }
      }
    }
    return repos;
  }
};
function activate(context) {
  const provider = new RepoProvider();
  vscode.window.registerTreeDataProvider("repoManagerView", provider);
  context.subscriptions.push(
    vscode.commands.registerCommand("repoManager.refresh", () => provider.refresh())
  );
  context.subscriptions.push(
    vscode.commands.registerCommand("repoManager.pullAll", async () => {
      const repos = await provider.getChildren();
      await vscode.window.withProgress(
        {
          location: vscode.ProgressLocation.Notification,
          title: "Pulling all repositories"
        },
        async () => {
          for (const repo of repos) {
            try {
              await runGitCommand(repo.repoPath, "git pull");
              vscode.window.showInformationMessage(`\u2705 Pulled ${repo.label}`);
            } catch (err) {
              vscode.window.showErrorMessage(`\u274C ${repo.label}: ${err}`);
            }
          }
          provider.refresh();
        }
      );
    })
  );
  context.subscriptions.push(
    vscode.commands.registerCommand("repoManager.fetchAll", async () => {
      const repos = await provider.getChildren();
      await vscode.window.withProgress(
        {
          location: vscode.ProgressLocation.Notification,
          title: "Fetching all repositories"
        },
        async () => {
          for (const repo of repos) {
            try {
              await runGitCommand(repo.repoPath, "git fetch");
              vscode.window.showInformationMessage(`\u2705 Fetched ${repo.label}`);
            } catch (err) {
              vscode.window.showErrorMessage(`\u274C ${repo.label}: ${err}`);
            }
          }
          provider.refresh();
        }
      );
    })
  );
}
function deactivate() {
}
// Annotate the CommonJS export names for ESM import in node:
0 && (module.exports = {
  activate,
  deactivate
});
//# sourceMappingURL=extension.js.map
