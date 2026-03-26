import { render, cleanup } from '@testing-library/svelte'
import { describe, it, expect, afterEach } from 'vitest'
import Icon from './Icon.svelte'
import { Home } from 'lucide-svelte'

describe('Icon component', () => {
  afterEach(() => cleanup())

  it('renders a lucide icon', () => {
    const { container } = render(Icon, { props: { icon: Home } })
    const svg = container.querySelector('svg')
    expect(svg).toBeTruthy()
  })

  it('applies custom size', () => {
    const { container } = render(Icon, { props: { icon: Home, size: 32 } })
    const svg = container.querySelector('svg')
    expect(svg?.getAttribute('width')).toBe('32')
    expect(svg?.getAttribute('height')).toBe('32')
  })

  it('applies custom class', () => {
    const { container } = render(Icon, { props: { icon: Home, class: 'text-red-500' } })
    const svg = container.querySelector('svg')
    expect(svg?.classList.contains('text-red-500')).toBe(true)
  })

  it('applies custom strokeWidth', () => {
    const { container } = render(Icon, { props: { icon: Home, strokeWidth: 3 } })
    const svg = container.querySelector('svg')
    expect(svg?.getAttribute('stroke-width')).toBe('3')
  })
})
