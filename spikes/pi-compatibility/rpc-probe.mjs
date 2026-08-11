import fs from "node:fs";
import os from "node:os";
import path from "node:path";
import { spawn } from "node:child_process";

const cliPath = process.env.PI_CLI ?? path.join(
  process.env.APPDATA ?? "",
  "npm",
  "node_modules",
  "@earendil-works",
  "pi-coding-agent",
  "dist",
  "cli.js",
);
const logPath = process.env.RPC_LOG;
const stderrPath = process.env.RPC_STDERR_LOG;
const commandFile = process.env.RPC_COMMAND_FILE;
const commandText = commandFile
  ? fs.readFileSync(commandFile, "utf8")
  : process.env.RPC_COMMANDS ?? "[]";
const commands = JSON.parse(commandText);
const timeoutMs = Number(process.env.RPC_TIMEOUT_MS ?? "30000");
const forceShellPath = process.env.RPC_FORCE_SHELL_PATH;

if (!logPath || !stderrPath || !Array.isArray(commands)) {
  throw new Error("RPC_LOG, RPC_STDERR_LOG, and RPC_COMMANDS are required");
}

fs.mkdirSync(path.dirname(logPath), { recursive: true });
const records = [];
const stderrChunks = [];
let stdoutBuffer = "";
let commandIndex = 0;
let finished = false;
let timer;
let cleanConfigDir;

function record(direction, message) {
  records.push({
    timestamp: new Date().toISOString(),
    direction,
    message,
  });
}

function cleanPathWithoutBash() {
  const original = process.env.Path ?? process.env.PATH ?? "";
  const entries = original.split(path.delimiter);
  const kept = entries.filter((entry) => {
    const candidate = entry.trim().replace(/^"|"$/g, "");
    if (!candidate) return false;
    if (/\\git\\(bin|usr\\bin)$/i.test(candidate)) return false;
    return !fs.existsSync(path.join(candidate, "bash.exe"));
  });
  return {
    value: kept.join(path.delimiter),
    originalCount: entries.filter(Boolean).length,
    removedCount: entries.filter(Boolean).length - kept.length,
    remainingBashExecutables: kept.filter((entry) => fs.existsSync(path.join(entry, "bash.exe"))).length,
  };
}

function finish(code, signal, error) {
  if (finished) return;
  finished = true;
  clearTimeout(timer);
  if (error) record("probe_error", { message: error.message });
  fs.writeFileSync(logPath, records.map((entry) => JSON.stringify(entry)).join("\n") + "\n", "utf8");
  fs.writeFileSync(stderrPath, stderrChunks.join(""), "utf8");
  if (cleanConfigDir) {
    fs.rmSync(cleanConfigDir, { recursive: true, force: true });
  }
  process.stderr.write(JSON.stringify({ exitCode: code, signal, error: error?.message ?? null }) + "\n");
  if (error || code !== 0) process.exitCode = 1;
}

function expectedResponse(response, command) {
  if (!command) return false;
  if (response?.type !== "response") return false;
  if (command.id !== undefined && response.id !== command.id) return false;
  return response.command === command.type;
}

function sendNext(child) {
  if (finished) return;
  if (commandIndex >= commands.length) {
    setTimeout(() => child.stdin.end(), 250);
    return;
  }
  const spec = commands[commandIndex++];
  const { waitFor, ...command } = spec;
  record("out", command);
  child.stdin.write(JSON.stringify(command) + "\n");
  child.__waitingFor = command;
  child.__waitForEvent = waitFor;
}

const childEnv = { ...process.env };
if (process.env.RPC_CLEAN_ENV === "1" || forceShellPath) {
  cleanConfigDir = fs.mkdtempSync(path.join(os.tmpdir(), "pi-gate0-"));
  childEnv.PI_CODING_AGENT_DIR = cleanConfigDir;
  if (forceShellPath) {
    fs.writeFileSync(
      path.join(cleanConfigDir, "settings.json"),
      JSON.stringify({ shellPath: forceShellPath }),
      "utf8",
    );
  }
}
if (process.env.RPC_CLEAN_ENV === "1") {
  const cleanPath = cleanPathWithoutBash();
  childEnv.Path = cleanPath.value;
  childEnv.PATH = childEnv.Path;
  childEnv.ProgramFiles = cleanConfigDir;
  childEnv["ProgramFiles(x86)"] = cleanConfigDir;
  record("probe_environment", {
    cleanEnv: true,
    originalPathEntries: cleanPath.originalCount,
    removedPathEntries: cleanPath.removedCount,
    remainingBashExecutables: cleanPath.remainingBashExecutables,
    isolatedConfig: true,
    overriddenKnownGitBashRoots: true,
    forcedShellPath: forceShellPath ?? null,
  });
}

const child = spawn(process.execPath, [cliPath, ...process.argv.slice(2)], {
  cwd: process.cwd(),
  env: childEnv,
  stdio: ["pipe", "pipe", "pipe"],
  windowsHide: true,
});

child.on("error", (error) => finish(null, null, error));
child.stdout.on("data", (chunk) => {
  stdoutBuffer += chunk.toString("utf8");
  let newlineIndex;
  while ((newlineIndex = stdoutBuffer.indexOf("\n")) >= 0) {
    let line = stdoutBuffer.slice(0, newlineIndex);
    stdoutBuffer = stdoutBuffer.slice(newlineIndex + 1);
    if (line.endsWith("\r")) line = line.slice(0, -1);
    if (!line) continue;
    let message;
    try {
      message = JSON.parse(line);
    } catch {
      message = { raw: line };
    }
    record("in", message);
    if (expectedResponse(message, child.__waitingFor)) {
      child.__waitingFor = undefined;
      if (!child.__waitForEvent) sendNext(child);
    }
    if (child.__waitForEvent && message.type === child.__waitForEvent) {
      child.__waitForEvent = undefined;
      sendNext(child);
    }
  }
});
child.stderr.on("data", (chunk) => stderrChunks.push(chunk.toString("utf8")));
child.on("close", (code, signal) => finish(code, signal, null));

timer = setTimeout(() => {
  record("probe_error", { message: `timed out after ${timeoutMs}ms` });
  child.kill();
  setTimeout(() => finish(null, "SIGTERM", new Error("RPC probe timeout")), 1000);
}, timeoutMs);

setTimeout(() => sendNext(child), 100);
