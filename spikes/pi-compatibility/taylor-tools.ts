import { execFile } from "node:child_process";
import { promises as fs } from "node:fs";
import path from "node:path";
import { promisify } from "node:util";
import type { ExtensionAPI, ExtensionContext } from "@earendil-works/pi-coding-agent";
import { Type } from "typebox";

const execFileAsync = promisify(execFile);

const taylorrunPath = process.env.TAYLORRUN_EXE;

function resolveWorkspacePath(ctx: ExtensionContext, requestedPath: string): string {
  const root = path.resolve(ctx.cwd);
  const resolved = path.resolve(root, requestedPath);
  const relative = path.relative(root, resolved);
  if (relative.startsWith("..") || path.isAbsolute(relative)) {
    throw new Error(`path escapes spike workspace: ${requestedPath}`);
  }
  return resolved;
}

const readTool = {
  name: "taylor_read",
  label: "Taylor read",
  description: "Read a text file through Taylor's explicit tool boundary.",
  parameters: Type.Object({
    path: Type.String({ description: "Path relative to the current workspace" }),
  }),
  async execute(_toolCallId: string, params: { path: string }, _signal: AbortSignal | undefined, _onUpdate: unknown, ctx: ExtensionContext) {
    const filePath = resolveWorkspacePath(ctx, params.path);
    const content = await fs.readFile(filePath, "utf8");
    return {
      content: [{ type: "text", text: content }],
      details: { tool: "taylor_read", path: params.path, bytes: Buffer.byteLength(content, "utf8") },
    };
  },
};

const writeTool = {
  name: "taylor_write",
  label: "Taylor write",
  description: "Write a text file through Taylor's explicit tool boundary.",
  parameters: Type.Object({
    path: Type.String({ description: "Path relative to the current workspace" }),
    content: Type.String({ description: "Complete replacement file content" }),
  }),
  async execute(_toolCallId: string, params: { path: string; content: string }, _signal: AbortSignal | undefined, _onUpdate: unknown, ctx: ExtensionContext) {
    const filePath = resolveWorkspacePath(ctx, params.path);
    await fs.writeFile(filePath, params.content, "utf8");
    return {
      content: [{ type: "text", text: `wrote ${params.path}` }],
      details: { tool: "taylor_write", path: params.path, bytes: Buffer.byteLength(params.content, "utf8") },
    };
  },
};

const runTool = {
  name: "taylor_run",
  label: "Taylor run",
  description: "Run a PowerShell 7 command through Taylor's explicit tool boundary.",
  parameters: Type.Object({
    command: Type.String({ description: "PowerShell command to run in the current workspace" }),
  }),
  async execute(_toolCallId: string, params: { command: string }, signal: AbortSignal | undefined, _onUpdate: unknown, ctx: ExtensionContext) {
    if (taylorrunPath) {
      const hostResult = await execFileAsync(taylorrunPath, [
        "-command",
        params.command,
        "-workdir",
        ctx.cwd,
      ], {
        cwd: ctx.cwd,
        encoding: "utf8",
        maxBuffer: 1024 * 1024,
        signal,
      });
      const payload = JSON.parse(String(hostResult.stdout).trim());
      if (payload.error) {
        throw new Error(`${payload.errorCode ?? "taylorrun"}: ${payload.error}`);
      }
      return {
        content: [{ type: "text", text: payload.stdout + payload.stderr || "(no output)" }],
        details: {
          tool: "taylor_run",
          transport: "go-host",
          executable: taylorrunPath,
          command: params.command,
          exitCode: payload.exitCode,
        },
      };
    }

    const result = await execFileAsync("pwsh", ["-NoProfile", "-NonInteractive", "-Command", params.command], {
      cwd: ctx.cwd,
      encoding: "utf8",
      maxBuffer: 1024 * 1024,
      signal,
    });
    return {
      content: [{ type: "text", text: `${result.stdout}${result.stderr}` || "(no output)" }],
      details: { tool: "taylor_run", command: params.command, exitCode: 0 },
    };
  },
};

export default function registerTaylorTools(pi: ExtensionAPI): void {
  pi.registerTool(readTool);
  pi.registerTool(writeTool);
  pi.registerTool(runTool);
}
