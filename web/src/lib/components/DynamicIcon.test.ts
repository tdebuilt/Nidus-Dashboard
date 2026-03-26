import { render, cleanup } from '@testing-library/svelte'
import { describe, it, expect, afterEach } from 'vitest'
import DynamicIcon from './DynamicIcon.svelte'

describe('DynamicIcon', () => {
  afterEach(() => {
    cleanup()
  })

  it('renders a known icon by name', () => {
    const { container } = render(DynamicIcon, { props: { name: 'House' } })
    const svg = container.querySelector('svg.lucide-icon')
    expect(svg).toBeTruthy()
    expect(svg?.classList.contains('lucide-house')).toBe(true)
  })

  it('renders fallback SVG for unknown icon name', () => {
    const { container } = render(DynamicIcon, { props: { name: 'NonExistentIcon999' } })
    const svg = container.querySelector('svg.lucide-fallback')
    expect(svg).toBeTruthy()
  })

  it('applies custom size', () => {
    const { container } = render(DynamicIcon, { props: { name: 'House', size: 32 } })
    const svg = container.querySelector('svg.lucide-icon')
    expect(svg?.getAttribute('width')).toBe('32')
    expect(svg?.getAttribute('height')).toBe('32')
  })

  it('applies custom color', () => {
    const { container } = render(DynamicIcon, { props: { name: 'House', color: '#ff0000' } })
    const svg = container.querySelector('svg.lucide-icon')
    expect(svg?.getAttribute('stroke')).toBe('#ff0000')
  })

  it('applies custom strokeWidth', () => {
    const { container } = render(DynamicIcon, { props: { name: 'House', strokeWidth: 3 } })
    const svg = container.querySelector('svg.lucide-icon')
    expect(svg?.getAttribute('stroke-width')).toBe('3')
  })

  it('uses default size of 24', () => {
    const { container } = render(DynamicIcon, { props: { name: 'House' } })
    const svg = container.querySelector('svg.lucide-icon')
    expect(svg?.getAttribute('width')).toBe('24')
    expect(svg?.getAttribute('height')).toBe('24')
  })

  it('uses currentColor as default color', () => {
    const { container } = render(DynamicIcon, { props: { name: 'House' } })
    const svg = container.querySelector('svg.lucide-icon')
    expect(svg?.getAttribute('stroke')).toBe('currentColor')
  })

  it('renders different icons correctly', () => {
    const { container: c1 } = render(DynamicIcon, { props: { name: 'Settings' } })
    expect(c1.querySelector('svg.lucide-settings')).toBeTruthy()
    cleanup()

    const { container: c2 } = render(DynamicIcon, { props: { name: 'LayoutDashboard' } })
    expect(c2.querySelector('svg.lucide-layout-dashboard')).toBeTruthy()
  })
})
