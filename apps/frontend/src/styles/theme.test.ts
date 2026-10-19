import { describe, it, expect } from 'vitest'
import { readdirSync, readFileSync, statSync } from 'node:fs'
import { join } from 'node:path'

const here = import.meta.dirname
const css = readFileSync(join(here, 'theme.css'), 'utf8')

// the text of one block, from its opening line to the closing brace at the start of a line
function block(opening: string): string {
  const start = css.indexOf(opening)
  expect(start, `${opening} in theme.css`).toBeGreaterThan(-1)
  return css.slice(start, css.indexOf('\n}', start))
}

function declared(text: string): Set<string> {
  return new Set([...text.matchAll(/^\s*(--[\w-]+):/gm)].map((m) => m[1]))
}

const info = declared(block('@theme {'))
const rootVars = declared(block(':root {'))
const neon = declared(block("[data-theme='neon'] {"))

describe('theme tokens', () => {
  it.each([
    '--color-bg', '--color-surface', '--color-raised', '--color-border',
    '--color-text', '--color-text-muted', '--color-text-faint',
    '--color-accent', '--color-accent-ink', '--color-accent-2',
    '--color-success', '--color-warning', '--color-danger',
    '--color-age-all', '--color-age-16', '--color-age-18', '--color-age-21',
    '--font-display', '--font-mono', '--font-body', '--radius-control',
  ])('defines %s for Info', (token) => {
    expect(info.has(token)).toBe(true)
  })

  it.each([
    '--color-tag-music', '--color-tag-comedy', '--color-tag-community', '--color-tag-art', '--color-tag-diy',
    '--color-tag-film', '--color-tag-literature', '--color-tag-sports', '--color-tag-free', '--color-tag-recurring',
  ])('defines the tag color %s in both themes', (token) => {
    expect(info.has(token)).toBe(true)
    expect(neon.has(token)).toBe(true)
  })

  it('overrides only tokens that exist for Info', () => {
    const known = new Set([...info, ...rootVars])
    const stray = [...neon].filter((token) => !known.has(token))
    expect(stray).toEqual([])
  })

  it('gives Neon its own fonts, glow, chamfer and uppercase labels', () => {
    for (const token of ['--font-display', '--font-mono', '--font-body', '--glow-accent', '--chamfer', '--label-transform']) {
      expect(neon.has(token), token).toBe(true)
    }
  })
})

function sources(dir: string): string[] {
  return readdirSync(dir).flatMap((name) => {
    const path = join(dir, name)
    if (statSync(path).isDirectory()) return sources(path)
    return /\.tsx$/.test(name) ? [path] : []
  })
}

// Components only use semantic token classes, so both themes can restyle them.
describe('components', () => {
  const src = join(here, '..')
  const palette = /\b(?:bg|text|border|ring|from|to|via|fill|stroke|divide|outline|shadow)-(?:red|green|blue|yellow|orange|pink|purple|sky|emerald|indigo|neutral|gray|slate|zinc|white|black)\b/
  const color = /#[0-9a-fA-F]{3,8}\b|\brgba?\(|\bhsla?\(/

  it.each(sources(src))('%s uses no palette class or literal color', (file) => {
    const text = readFileSync(file, 'utf8')
    expect(text.match(palette)?.[0]).toBeUndefined()
    expect(text.match(color)?.[0]).toBeUndefined()
  })
})
