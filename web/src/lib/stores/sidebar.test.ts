import { describe, it, expect, beforeEach } from 'vitest'
import { get } from 'svelte/store'
import { sidebarOpen, toggleSidebar, openSidebar, closeSidebar } from './sidebar'

describe('Sidebar store', () => {
  beforeEach(() => {
    sidebarOpen.set(false)
  })

  it('starts closed', () => {
    expect(get(sidebarOpen)).toBe(false)
  })

  it('toggles open', () => {
    toggleSidebar()
    expect(get(sidebarOpen)).toBe(true)
  })

  it('toggles closed', () => {
    toggleSidebar() // open
    toggleSidebar() // close
    expect(get(sidebarOpen)).toBe(false)
  })

  it('closes explicitly', () => {
    sidebarOpen.set(true)
    closeSidebar()
    expect(get(sidebarOpen)).toBe(false)
  })

  it('opens explicitly', () => {
    expect(get(sidebarOpen)).toBe(false)
    openSidebar()
    expect(get(sidebarOpen)).toBe(true)
  })
})
