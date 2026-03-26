import { readable } from 'svelte/store'

export type Breakpoint = 'mobile' | 'tablet' | 'desktop'

export const breakpoint = readable<Breakpoint>('desktop', (set) => {
  if (typeof window === 'undefined' || typeof window.matchMedia !== 'function') return

  const mqMobile = window.matchMedia('(max-width: 767px)')
  const mqTablet = window.matchMedia('(min-width: 768px) and (max-width: 1023px)')

  function update() {
    if (mqMobile.matches) set('mobile')
    else if (mqTablet.matches) set('tablet')
    else set('desktop')
  }

  update()
  mqMobile.addEventListener('change', update)
  mqTablet.addEventListener('change', update)

  return () => {
    mqMobile.removeEventListener('change', update)
    mqTablet.removeEventListener('change', update)
  }
})
