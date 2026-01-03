import * as vscode from "vscode";
import { exec } from "child_process";
import * as fs from "fs";
import * as path from "path";

// --------------------------
// Helper functions
// --------------------------
export function runGitCommand(repoPath: string, command: string): Promise<string> {
  return new Promise((resolve, reject) => {
    exec(command, { cwd: repoPath }, (err, stdout, stderr) => {
      if (err) {
        reject(stderr || err.message);
      } else {
        resolve(stdout.trim());
      }
    });
  });
}

export async function getRepoStatus(repoPath: string) {
  try {
    const branch = (await runGitCommand(repoPath, "git branch --show-current")) || "detached";
    const dirtyOutput = await runGitCommand(repoPath, "git status --porcelain");
    const dirty = dirtyOutput.length > 0;
    let ahead = 0, behind = 0;

    try {
      const revList = await runGitCommand(
        repoPath,
        "git rev-list --left-right --count HEAD...@{upstream}"
      );
      [ahead, behind] = revList.split("\t").map(Number);
    } catch { /* upstream may not exist */ }

    return { branch, dirty, ahead, behind };
  } catch {
    return { branch: "unknown", dirty: false, ahead: 0, behind: 0 };
  }
}

// --------------------------
// TreeItem for each repo
// --------------------------
class RepoItem extends vscode.TreeItem {
  constructor(
    public readonly repoPath: string,
    public readonly label: string,
    public readonly status?: string
  ) {
    super(label, vscode.TreeItemCollapsibleState.None);
    this.tooltip = repoPath;
    this.description = status;
    this.contextValue = "repo";
  }
}

// --------------------------
// TreeDataProvider
// --------------------------
class RepoProvider implements vscode.TreeDataProvider<RepoItem> {
  private _onDidChangeTreeData = new vscode.EventEmitter<RepoItem | undefined>();
  readonly onDidChangeTreeData = this._onDidChangeTreeData.event;

  refresh(): void {
    this._onDidChangeTreeData.fire(undefined);
  }

  getTreeItem(item: RepoItem): vscode.TreeItem {
    return item;
  }

  async getChildren(): Promise<RepoItem[]> {
    const folders = vscode.workspace.workspaceFolders;
    if (!folders) {return [];}

    const repos: RepoItem[] = [];
    const config = vscode.workspace.getConfiguration('rhizaManager');
    const repositoryRoot = config.get<string>('repositoryRoot', 'subfolders');

    for (const folder of folders) {
      const root = folder.uri.fsPath;
      
      // If repositoryRoot is 'workspace', check if the workspace folder itself is a repo
      if (repositoryRoot === 'workspace') {
        if (fs.existsSync(path.join(root, ".git"))) {
          const status = await getRepoStatus(root);
          const desc = `${status.branch} · ${status.dirty ? "dirty" : "clean"} · ↑${status.ahead} ↓${status.behind}`;
          const repoName = path.basename(root);
          repos.push(new RepoItem(root, repoName, desc));
        }
      } else {
        // Default behavior: look for repos in subfolders
        for (const entry of fs.readdirSync(root)) {
          const fullPath = path.join(root, entry);
          if (fs.statSync(fullPath).isDirectory() && fs.existsSync(path.join(fullPath, ".git"))) {
            const status = await getRepoStatus(fullPath);
            const desc = `${status.branch} · ${status.dirty ? "dirty" : "clean"} · ↑${status.ahead} ↓${status.behind}`;
            repos.push(new RepoItem(fullPath, entry, desc));
          }
        }
      }
    }

    return repos;
  }
}

// --------------------------
// Activate Extension
// --------------------------
export function activate(context: vscode.ExtensionContext) {
  const provider = new RepoProvider();
  vscode.window.registerTreeDataProvider("repoManagerView", provider);

  // Refresh command
  context.subscriptions.push(
    vscode.commands.registerCommand("repoManager.refresh", () => provider.refresh())
  );

  // Pull All command
  context.subscriptions.push(
    vscode.commands.registerCommand("repoManager.pullAll", async () => {
      const repos = await provider.getChildren();
      await vscode.window.withProgress(
        {
          location: vscode.ProgressLocation.Notification,
          title: "Pulling all repositories",
        },
        async () => {
          for (const repo of repos) {
            try {
              await runGitCommand(repo.repoPath, "git pull");
              vscode.window.showInformationMessage(`✅ Pulled ${repo.label}`);
            } catch (err) {
              vscode.window.showErrorMessage(`❌ ${repo.label}: ${err}`);
            }
          }
          provider.refresh();
        }
      );
    })
  );

  // Fetch All command
  context.subscriptions.push(
    vscode.commands.registerCommand("repoManager.fetchAll", async () => {
      const repos = await provider.getChildren();
      await vscode.window.withProgress(
        {
          location: vscode.ProgressLocation.Notification,
          title: "Fetching all repositories",
        },
        async () => {
          for (const repo of repos) {
            try {
              await runGitCommand(repo.repoPath, "git fetch");
              vscode.window.showInformationMessage(`✅ Fetched ${repo.label}`);
            } catch (err) {
              vscode.window.showErrorMessage(`❌ ${repo.label}: ${err}`);
            }
          }
          provider.refresh();
        }
      );
    })
  );
}

// --------------------------
// Deactivate Extension
// --------------------------
export function deactivate() {}
