#!/usr/bin/env node
/**
 * Fix invalid Playwright "text engine" comma lists.
 *
 * Problem
 * -------
 * These specs contain locators such as:
 *
 *     page.locator('text=概览, text=Overview, text=工作空间')
 *
 * Playwright treats the whole string as ONE selector. The `text=` engine is an
 * *engine* selector, not CSS, so a comma-joined list of `text=` engines is not a
 * valid selector union. Playwright instead interprets everything after the first
 * `text=` as a single literal text value ("概览, text=Overview, text=工作空间"),
 * which never matches anything on the page. The assertion then burns its full
 * expect timeout and fails. Mixed forms are worse: a leading CSS selector list
 * followed by `text=` parts (`[class*="status"], text=进行中`) is invalid CSS and
 * fails outright.
 *
 * A comma inside a plain CSS selector list IS valid (`h1, h2`), so only locators
 * that actually contain a `text=` engine are rewritten.
 *
 * Fix
 * ---
 * Rewrite the union into a real "any of" locator chain:
 *
 *     page.locator('text=概览').or(page.locator('text=Overview')).or(page.locator('text=工作空间'))
 *
 * Every comma-separated part keeps its own selector engine, so semantics are
 * preserved (match if ANY part matches), and `.first()` / `expect(...).toBeVisible()`
 * keep working.
 *
 * Usage:
 *   node scripts/fix-invalid-text-locators.mjs <dir-or-file> [...] [--dry-run]
 */

import fs from 'node:fs';
import path from 'node:path';

const args = process.argv.slice(2);
const dryRun = args.includes('--dry-run');
const targets = args.filter((a) => a !== '--dry-run');

if (targets.length === 0) {
  console.error('usage: node scripts/fix-invalid-text-locators.mjs <dir-or-file> [...] [--dry-run]');
  process.exit(2);
}

/** Collect .spec.ts files under a path (recursively), skipping node_modules. */
function collectSpecs(target) {
  const stat = fs.statSync(target);
  if (stat.isFile()) return [target];
  const out = [];
  for (const entry of fs.readdirSync(target, { withFileTypes: true })) {
    if (entry.name === 'node_modules' || entry.name === 'test-results' || entry.name === 'playwright-report') continue;
    const full = path.join(target, entry.name);
    if (entry.isDirectory()) out.push(...collectSpecs(full));
    else if (entry.isFile() && /\.spec\.ts$/.test(entry.name)) out.push(full);
  }
  return out;
}

/**
 * Split a Playwright selector string on top-level commas.
 * Commas inside single quotes, double quotes, or /regex/ literals are ignored.
 */
function splitTopLevel(selector) {
  const parts = [];
  let current = '';
  let quote = null;
  let inRegex = false;

  for (let i = 0; i < selector.length; i++) {
    const ch = selector[i];
    const prev = selector[i - 1];

    if (quote) {
      current += ch;
      if (ch === quote && prev !== '\\') quote = null;
      continue;
    }
    if (inRegex) {
      current += ch;
      if (ch === '/' && prev !== '\\') inRegex = false;
      continue;
    }
    if (ch === '"' || ch === "'") {
      quote = ch;
      current += ch;
      continue;
    }
    if (ch === '=' && current.endsWith('text')) {
      // entering a text= engine; a following /.../ is a regex payload
      // handled by the generic branch below once we see the slash
    }
    if (ch === '/' && /text=$/.test(current)) {
      inRegex = true;
      current += ch;
      continue;
    }
    if (ch === ',') {
      parts.push(current);
      current = '';
      continue;
    }
    current += ch;
  }
  parts.push(current);
  return parts.map((p) => p.trim()).filter((p) => p.length > 0);
}

const LOCATOR_RE = /(\b[A-Za-z_$][\w$]*)\.locator\(\s*(['"])((?:\\.|(?!\2)[^\\])*)\2\s*\)/g;

let filesChanged = 0;
let rewrites = 0;

for (const target of targets) {
  for (const file of collectSpecs(target)) {
    const original = fs.readFileSync(file, 'utf8');

    const updated = original.replace(LOCATOR_RE, (match, receiver, quote, body) => {
      // Only touch unions that involve the text engine.
      if (!body.includes('text=')) return match;
      const parts = splitTopLevel(body);
      if (parts.length < 2) return match;

      // Left-associative chain: A.or(B).or(C) — same match set as A.or(B.or(C))
      // but reads naturally and keeps `.first()` semantics on the whole union.
      let rebuilt = `${receiver}.locator(${quote}${parts[0]}${quote})`;
      for (let i = 1; i < parts.length; i++) {
        rebuilt += `.or(${receiver}.locator(${quote}${parts[i]}${quote}))`;
      }

      rewrites++;
      return rebuilt;
    });

    if (updated !== original) {
      filesChanged++;
      if (dryRun) {
        const before = original.split('\n').filter((l) => /text=[^']*,\s*text=/.test(l)).length;
        console.log(`[dry-run] ${path.relative(process.cwd(), file)}  (~${before} lines)`);
      } else {
        fs.writeFileSync(file, updated, 'utf8');
        console.log(`[fixed]   ${path.relative(process.cwd(), file)}`);
      }
    }
  }
}

console.log(`\n${dryRun ? 'would rewrite' : 'rewrote'} ${rewrites} locator(s) in ${filesChanged} file(s)`);
