/**
 * Locale negotiation.
 *
 * The UI ships two locales. On the first visit there is no stored preference,
 * so the browser's own language decides; after that the switcher's explicit
 * choice wins and is persisted by `setLocale`.
 *
 * A regional tag that we do not translate (zh-TW, en-GB, en-IN) maps to the
 * closest language we have rather than falling through to the default: a
 * traditional-Chinese reader gets simplified Chinese, which is far closer than
 * English would be.
 */

export type Locale = 'zh-CN' | 'en-US'

export const DEFAULT_LOCALE: Locale = 'zh-CN'

export const SUPPORTED_LOCALES: readonly Locale[] = ['zh-CN', 'en-US']

/** Map any BCP-47 tag to one of the shipped locales, or null if unrelated. */
export function matchLocale(tag: string | null | undefined): Locale | null {
  if (!tag) return null
  const primary = tag.trim().toLowerCase().split(/[-_]/)[0]
  if (primary === 'zh') return 'zh-CN'
  if (primary === 'en') return 'en-US'
  return null
}

/**
 * Pick the locale to start with.
 *
 * Order: valid stored preference, then the browser's ordered language list,
 * then the default.
 */
export function resolveInitialLocale(
  stored: string | null | undefined,
  browserLanguages: readonly string[] = [],
  fallback: Locale = DEFAULT_LOCALE,
): Locale {
  const fromStorage = matchLocale(stored)
  if (fromStorage) return fromStorage

  for (const tag of browserLanguages) {
    const match = matchLocale(tag)
    if (match) return match
  }

  return fallback
}

/** `navigator.languages`, guarded for non-browser (test/SSR) environments. */
export function browserLanguages(): string[] {
  if (typeof navigator === 'undefined') return []
  const ordered = navigator.languages
  if (ordered && ordered.length > 0) return [...ordered]
  return navigator.language ? [navigator.language] : []
}
