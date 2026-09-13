#!/usr/bin/env node
/**
 * Aggregate a Playwright JSON report into an end-to-end coverage summary.
 *
 * Usage:
 *   node scripts/e2e-coverage-report.mjs <report.json> [--label "tests/"] [--json-out agg.json]
 *
 * Emits a markdown section: totals, a per-suite-directory table, and the list
 * of failing cases with their retry status.
 */

import fs from 'node:fs';
import path from 'node:path';

const args = process.argv.slice(2);
const reportPath = args.find((a) => !a.startsWith('--'));
const label = args.includes('--label') ? args[args.indexOf('--label') + 1] : 'E2E';
const jsonOut = args.includes('--json-out') ? args[args.indexOf('--json-out') + 1] : null;

if (!reportPath) {
  console.error('usage: node scripts/e2e-coverage-report.mjs <report.json> [--label X] [--json-out f]');
  process.exit(2);
}

const report = JSON.parse(fs.readFileSync(reportPath, 'utf8'));

/** Walk the suite tree, collecting every spec with the file it belongs to. */
function collect(suites, file, out) {
  for (const suite of suites ?? []) {
    const suiteFile = suite.file ?? file ?? '(unknown)';
    for (const spec of suite.specs ?? []) {
      const results = spec.tests?.flatMap((t) => t.results ?? []) ?? [];
      // A spec's outcome ignores intermediate failures that a retry recovered.
      const finalStatus = spec.ok ? 'passed' : (spec.tests?.[0]?.status ?? 'failed');
      const attempted = results.filter((r) => r.status !== 'skipped');
      const retried = attempted.length > 1;
      const recoveredOnRetry = !spec.ok ? false : attempted.some((r) => r.status === 'failed');
      out.push({
        file: suiteFile.replace(/\\/g, '/'),
        title: spec.title,
        ok: !!spec.ok,
        status: finalStatus,
        skipped: results.length > 0 && results.every((r) => r.status === 'skipped'),
        retried,
        recoveredOnRetry,
        durationMs: results.reduce((a, r) => a + (r.duration ?? 0), 0),
      });
    }
    collect(suite.suites, suiteFile, out);
  }
}

const specs = [];
collect(report.suites, null, specs);

// Group by the first path segment (tests/ layout: <NN-module>/file.spec.ts)
const byDir = new Map();
for (const s of specs) {
  const dir = s.file.includes('/') ? s.file.split('/')[0] : '(root)';
  if (!byDir.has(dir)) byDir.set(dir, { passed: 0, failed: 0, skipped: 0, flaky: 0, total: 0, ms: 0 });
  const b = byDir.get(dir);
  b.total++;
  b.ms += s.durationMs;
  if (s.skipped) b.skipped++;
  else if (s.ok) b.passed++;
  else b.failed++;
  if (s.recoveredOnRetry) b.flaky++;
}

const totals = {
  total: specs.length,
  passed: specs.filter((s) => s.ok && !s.skipped).length,
  failed: specs.filter((s) => !s.ok && !s.skipped).length,
  skipped: specs.filter((s) => s.skipped).length,
  flaky: specs.filter((s) => s.recoveredOnRetry).length,
  durationMs: specs.reduce((a, s) => a + s.durationMs, 0),
};

const failures = specs.filter((s) => !s.ok && !s.skipped);

const mmss = (ms) => `${Math.floor(ms / 60000)}m ${String(Math.round((ms % 60000) / 1000)).padStart(2, '0')}s`;
const rate = totals.passed + totals.failed > 0
  ? ((totals.passed / (totals.passed + totals.failed)) * 100).toFixed(1)
  : '0.0';

let md = '';
md += `### ${label}\n\n`;
md += `| 指标 | 数值 |\n|------|------|\n`;
md += `| 用例总数 | ${totals.total} |\n`;
md += `| 通过 | ${totals.passed} |\n`;
md += `| 失败 | ${totals.failed} |\n`;
md += `| 跳过 | ${totals.skipped} |\n`;
md += `| 通过率 | ${rate}% |\n`;
md += `| 重试后通过（flaky） | ${totals.flaky} |\n`;
md += `| 执行耗时 | ${mmss(totals.durationMs)} |\n\n`;

md += `| 模块目录 | 总数 | 通过 | 失败 | 跳过 | flaky | 耗时 |\n|---|---|---|---|---|---|---|\n`;
for (const [dir, b] of [...byDir.entries()].sort()) {
  md += `| ${dir} | ${b.total} | ${b.passed} | ${b.failed} | ${b.skipped} | ${b.flaky} | ${mmss(b.ms)} |\n`;
}
md += '\n';

if (failures.length) {
  md += `#### 失败用例清单 (${failures.length})\n\n`;
  md += `| 文件 | 用例 | 重试 |\n|---|---|---|\n`;
  for (const f of failures.sort((a, b) => a.file.localeCompare(b.file))) {
    md += `| ${f.file} | ${f.title} | ${f.retried ? '有' : '无'} |\n`;
  }
  md += '\n';
}

const flakyList = specs.filter((s) => s.recoveredOnRetry);
if (flakyList.length) {
  md += `#### 不稳定用例（首跑失败、重试通过）(${flakyList.length})\n\n`;
  for (const f of flakyList.sort((a, b) => a.file.localeCompare(b.file))) {
    md += `- \`${f.file}\` — ${f.title}\n`;
  }
  md += '\n';
}

console.log(md);

if (jsonOut) {
  fs.writeFileSync(jsonOut, JSON.stringify({ label, totals, byDir: Object.fromEntries(byDir), failures, flaky: flakyList }, null, 2), 'utf8');
  console.error(`aggregate written to ${jsonOut}`);
}
