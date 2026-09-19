import { test, expect } from "@playwright/test";
import { _electron as electron } from "playwright";

import fs from "fs";
import os from "os";
import path from "path";

const DEFAULT_REPO_ID = "repository";

let electronApp;
let mainPath;
let executablePath;
let tmpAppDataDir;
const deferredAppDataCleanup = new Set();

function getGoreeCloudBackupDir() {
  switch (process.platform + "/" + process.arch) {
    case "darwin/x64":
      return path.resolve("../dist/kopia-ui/mac");
    case "darwin/arm64":
      return path.resolve("../dist/kopia-ui/mac-arm64");
    case "linux/x64":
      // on Linux we must run from installed location due to AppArmor profile
      return path.resolve("/opt/GoreeCloud Backup");
    case "linux/arm64":
      // on Linux we must run from installed location due to AppArmor profile
      return path.resolve("/opt/GoreeCloud Backup");
    case "win32/x64":
      return path.resolve("../dist/kopia-ui/win-unpacked");
    default:
      return null;
  }
}

function getMainPath(appDir) {
  switch (process.platform) {
    case "darwin":
      return path.join(
        appDir,
        "GoreeCloud Backup.app",
        "Contents",
        "Resources",
        "app.asar",
        "public",
        "electron.js",
      );
    default:
      return path.join(
        appDir,
        "resources",
        "app.asar",
        "public",
        "electron.js",
      );
  }
}

function getExecutablePath(appDir) {
  switch (process.platform) {
    case "win32":
      return path.join(appDir, "GoreeCloud Backup.exe");
    case "darwin":
      return path.join(
        appDir,
        "GoreeCloud Backup.app",
        "Contents",
        "MacOS",
        "GoreeCloud Backup",
      );
    default:
      // Linux intentionally retains the inherited internal executable name for
      // package/AppArmor compatibility while the installed product identity is GoreeCloud Backup.
      return path.join(appDir, "kopia-ui");
  }
}

/**
 * Creates a temporary application data directory along with the kopia
 * directory for testing purposes.
 *
 * @returns {string} The path to the created temporary directory.
 */
function createTemporaryAppDataDir() {
  const tmpDir = fs.mkdtempSync(
    path.join(os.tmpdir(), "goreecloud-backup-test-"),
  );
  fs.mkdirSync(path.join(tmpDir, "kopia"));
  return tmpDir;
}

/**
 * Launches a new instance of the Electron app with the given app data directory.
 *
 * Also captures page errors and console messages and logs them to the console.
 *
 * @param {string} appDataDir - the path to the app data directory
 * @returns {Promise<Electron.App>} - a promise that resolves to the launched app
 */
async function launchApp(appDataDir) {
  const app = await electron.launch({
    args: [mainPath],
    executablePath: executablePath,
    env: {
      ...process.env,
      KOPIA_CUSTOM_APPDATA: appDataDir,
    },
  });

  app.on("window", async (page) => {
    const filename = page.url()?.split("/").pop();
    console.log(`Window opened: ${filename}`);

    page.on("pageerror", (error) => {
      console.error(error);
    });
    page.on("console", (msg) => {
      console.log(msg.text());
    });
  });

  return app;
}

/**
 * Waits for the embedded backup engine to start up by delaying for a specified duration.
 *
 * @returns {Promise<void>} A promise that resolves after the delay.
 */
function waitForBackupEngineToStartup() {
  return new Promise((resolve) => {
    setTimeout(() => {
      resolve();
    }, 2500);
  });
}

/**
 * Waits until the packaged Electron child process has actually exited.
 *
 * Playwright may finish its close request before Windows releases every handle
 * beneath the temporary application profile. Waiting for the child-process
 * exit/close signal keeps test cleanup from racing those handle releases.
 *
 * @param {import("child_process").ChildProcess} childProcess
 * @param {number} timeoutMs
 * @returns {Promise<void>}
 */
function waitForElectronProcessExit(childProcess, timeoutMs = 10000) {
  if (childProcess.exitCode !== null || childProcess.signalCode !== null) {
    return Promise.resolve();
  }

  return new Promise((resolve, reject) => {
    let timeoutID;

    const cleanup = () => {
      clearTimeout(timeoutID);
      childProcess.off("exit", onExit);
      childProcess.off("close", onExit);
    };

    const onExit = () => {
      cleanup();
      resolve();
    };

    timeoutID = setTimeout(() => {
      cleanup();
      reject(new Error("timed out waiting for packaged Electron process exit"));
    }, timeoutMs);

    childProcess.once("exit", onExit);
    childProcess.once("close", onExit);
  });
}

/**
 * Closes the isolated packaged Electron test process without allowing teardown
 * to consume the test timeout. The embedded test servers receive an immediate
 * stop request, then the test-only Electron process is terminated and its
 * OS-level exit is observed before the temporary profile is removed.
 *
 * @param {Electron.App} app
 * @returns {Promise<void>}
 */
async function closePackagedApp(app) {
  const packagedProcess = app.process();

  // Shut down the embedded backup-engine children first. A forced Electron
  // termination before these children exit can orphan them on Windows and keep
  // handles open beneath the temporary application profile.
  let stopRequestTimeoutID;
  try {
    await Promise.race([
      app.evaluate(async ({ app: electronApplication }) => {
        await electronApplication.testHooks.stopAllServers();
      }),
      new Promise((_, reject) => {
        stopRequestTimeoutID = setTimeout(() => {
          reject(new Error("timed out waiting for embedded server shutdown"));
        }, 7000);
      }),
    ]);
  } catch (error) {
    console.warn("graceful embedded server shutdown failed:", error.message);

    try {
      await app.evaluate(async ({ app: electronApplication }) => {
        electronApplication.testHooks.stopAllServersNow();
      });
    } catch (fallbackError) {
      console.warn(
        "unable to request fallback embedded server shutdown:",
        fallbackError.message,
      );
    }
  } finally {
    clearTimeout(stopRequestTimeoutID);
  }

  let gracefulCloseTimeoutID;
  try {
    await Promise.race([
      app.close(),
      new Promise((_, reject) => {
        gracefulCloseTimeoutID = setTimeout(() => {
          reject(new Error("timed out waiting for graceful Electron close"));
        }, 3000);
      }),
    ]);
  } catch (error) {
    console.warn("graceful Electron close failed:", error.message);
  } finally {
    clearTimeout(gracefulCloseTimeoutID);
  }

  if (
    packagedProcess.exitCode === null &&
    packagedProcess.signalCode === null
  ) {
    packagedProcess.kill("SIGKILL");
  }

  await waitForElectronProcessExit(packagedProcess, 5000);

  // Windows can release application-profile handles shortly after the complete
  // Electron process tree is torn down. Keep this bounded and small.
  if (process.platform === "win32") {
    await new Promise((resolve) => setTimeout(resolve, 500));
  }
}

test.beforeAll(() => {
  const appDir = getGoreeCloudBackupDir();
  expect(appDir).not.toBeNull();

  mainPath = getMainPath(appDir);
  executablePath = getExecutablePath(appDir);

  console.log("main path", mainPath);
  console.log("executable path", executablePath);

  process.env.CI = "e2e";
  process.env.KOPIA_UI_TESTING = "1";
});

test.beforeEach(async () => {
  electronApp = undefined;
  tmpAppDataDir = createTemporaryAppDataDir();
});

test.afterEach(async () => {
  if (electronApp) {
    await closePackagedApp(electronApp);
  }

  try {
    fs.rmSync(tmpAppDataDir, {
      recursive: true,
      force: true,
      maxRetries: 4,
      retryDelay: 250,
    });
  } catch (error) {
    if (!["EPERM", "EBUSY", "ENOTEMPTY"].includes(error.code)) {
      throw error;
    }

    // Electron/Chromium can release profile handles shortly after the main
    // process exits, especially on Windows. Defer only transient lock cleanup;
    // afterAll still fails if the directory cannot be removed.
    deferredAppDataCleanup.add(tmpAppDataDir);
    console.warn("deferring locked application profile cleanup:", error.message);
  }
});

test.afterAll(async ({}, testInfo) => {
  testInfo.setTimeout(60000);

  for (const appDataDir of deferredAppDataCleanup) {
    fs.rmSync(appDataDir, {
      recursive: true,
      force: true,
      maxRetries: 40,
      retryDelay: 250,
    });
  }

  deferredAppDataCleanup.clear();
});

test("opens repository window on first start", async () => {
  electronApp = await launchApp(tmpAppDataDir);

  await electronApp.evaluate(async ({ app }) => {
    app.testHooks.showRepoWindow("repository");
  });

  const page = await electronApp.firstWindow();
  expect(page).toBeTruthy();

  const appVersion = await electronApp.evaluate(async ({ app }) =>
    app.getVersion(),
  );
  const nativeWindowTitles = await electronApp.evaluate(async ({ app }) =>
    app.testHooks.repositoryWindowTitles(),
  );

  expect(nativeWindowTitles).toContain(`GoreeCloud Backup v${appVersion}`);

  await electronApp.evaluate(async ({ app }) => {
    return app.testHooks.tray.popUpContextMenu();
  });

  await electronApp.evaluate(async ({ app }) => {
    return app.testHooks.tray.closeContextMenu();
  });
});

test("adds default repository if no repository is configured", async () => {
  electronApp = await launchApp(tmpAppDataDir);

  await waitForBackupEngineToStartup();

  const configs = await electronApp.evaluate(async ({ app }) => {
    return app.testHooks.allConfigs();
  });
  expect(configs).toStrictEqual([DEFAULT_REPO_ID]);
});

test("doesn't open repository window if the default repository config exists", async () => {
  fs.writeFileSync(
    path.join(tmpAppDataDir, "kopia", `${DEFAULT_REPO_ID}.config`),
    "",
  );

  electronApp = await launchApp(tmpAppDataDir);

  await waitForBackupEngineToStartup();
  const windows = electronApp.windows();
  expect(windows).toHaveLength(0);
});

test.describe("when non-default repository config exists", () => {
  const NON_DEFAULT_REPO_ID = "repository-42";

  test.beforeEach(async () => {
    fs.writeFileSync(
      path.join(tmpAppDataDir, "kopia", `${NON_DEFAULT_REPO_ID}.config`),
      "",
    );
  });

  test("doesn't open repository window if non-default repository config exists", async () => {
    electronApp = await launchApp(tmpAppDataDir);

    await waitForBackupEngineToStartup();
    const windows = electronApp.windows();
    expect(windows).toHaveLength(0);
  });

  test("doesn't add default repository", async () => {
    electronApp = await launchApp(tmpAppDataDir);

    await waitForBackupEngineToStartup();

    const configs = await electronApp.evaluate(async ({ app }) => {
      return app.testHooks.allConfigs();
    });
    expect(configs).toStrictEqual([NON_DEFAULT_REPO_ID]);
  });
});
