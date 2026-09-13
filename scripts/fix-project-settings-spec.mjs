// One-off repair for tests/03-project/settings.spec.ts
//
// Two independent defects made all 20 tests in this file fail:
//
//  1. Wrong URL. The suite navigated to
//        /workspace/qa-test/project/2347?tab=settings
//     Project.vue only redirects to the settings page from a *watcher* on
//     activeTab, and the watcher never fires when 'settings' is the initial
//     value from the query string. The page therefore rendered the project
//     tabs with no tab content. The real settings route is .../settings
//     (views/ProjectSettings.vue).
//
//  2. Wrong assertion text. Every test asserted `text=项目设置` / `text=Settings`,
//     but views/ProjectSettings.vue renders its title from the active section
//     label (default t('settings.overview') = "概览") and never contains the
//     literal string "项目设置". The stable marker of the settings page is its
//     subtitle t('settings.configureProject'), which is rendered in every
//     section.

import fs from 'node:fs';

const file = 'tests/03-project/settings.spec.ts';
let src = fs.readFileSync(file, 'utf8');
const original = src;

// 1) Point at the real settings route.
src = src.split("'/workspace/qa-test/project/2347?tab=settings'")
         .join("'/workspace/qa-test/project/2347/settings'");

// 2) Assert on the settings page's own subtitle instead of text that never renders.
const oldAssert = "page.locator('text=项目设置').or(page.locator('text=Settings'))";
const newAssert =
  "page.locator('text=配置项目级设置').or(page.locator('text=Configure project-level settings'))";

const replaced = src.split(oldAssert).length - 1;
src = src.split(oldAssert).join(newAssert);

fs.writeFileSync(file, src, 'utf8');

console.log(`assertions replaced: ${replaced}`);
console.log(`url rewritten:       ${original !== src && src.includes("project/2347/settings")}`);
console.log(`residual '项目设置': ${src.split('项目设置').length - 1}`);
console.log(`residual '?tab=settings': ${src.split('?tab=settings').length - 1}`);
