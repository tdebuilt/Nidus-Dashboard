import { render, screen, cleanup } from '@testing-library/svelte'
import { describe, it, expect, afterEach } from 'vitest'
import IconPicker from './IconPicker.svelte'

describe('IconPicker', () => {
  afterEach(() => cleanup())

  it('renders the icon picker grid', () => {
    render(IconPicker)
    expect(screen.getByTestId('icon-picker')).toBeTruthy()
  })

  it('shows 24 icon options', () => {
    render(IconPicker)
    const buttons = screen.getAllByRole('button')
    expect(buttons.length).toBe(24)
  })

  it('has folder icon by default', () => {
    render(IconPicker)
    expect(screen.getByTestId('icon-option-folder')).toBeTruthy()
  })

  it('shows known icon options', () => {
    render(IconPicker)
    expect(screen.getByTestId('icon-option-server')).toBeTruthy()
    expect(screen.getByTestId('icon-option-monitor')).toBeTruthy()
    expect(screen.getByTestId('icon-option-shield')).toBeTruthy()
  })
})
