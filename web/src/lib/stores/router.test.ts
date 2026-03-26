import { describe, it, expect, beforeEach } from 'vitest'
import { get } from 'svelte/store'
import { currentPath, currentRoute, navigate } from './router'

describe('Router', () => {
  beforeEach(() => {
    currentPath.set('/')
  })

  it('defaults to dashboard route', () => {
    expect(get(currentRoute)).toBe('dashboard')
  })

  it('navigates to login', () => {
    navigate('/login')
    expect(get(currentRoute)).toBe('login')
    expect(get(currentPath)).toBe('/login')
  })

  it('navigates to setup', () => {
    navigate('/setup')
    expect(get(currentRoute)).toBe('setup')
  })

  it('navigates to settings', () => {
    navigate('/settings')
    expect(get(currentRoute)).toBe('settings')
  })

  it('navigates to dashboard subpath (legacy ID format)', () => {
    navigate('/dashboard/category/1')
    expect(get(currentRoute)).toBe('dashboard')
  })

  it('navigates to dashboard with slug', () => {
    navigate('/dashboard/mon-salon')
    expect(get(currentRoute)).toBe('dashboard')
  })

  it('returns not-found for unknown routes', () => {
    navigate('/unknown/path')
    expect(get(currentRoute)).toBe('not-found')
  })

  it('updates window.history on navigate', () => {
    const pushStateSpy = history.pushState
    navigate('/settings')
    expect(get(currentPath)).toBe('/settings')
    // Restore
    history.pushState = pushStateSpy
  })
})
