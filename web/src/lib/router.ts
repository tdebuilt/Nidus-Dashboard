import { writable, derived, get } from 'svelte/store';

export type Route = 'login' | 'setup' | 'dashboard' | 'settings' | 'not-found';

const PUBLIC_ROUTES: Route[] = ['login', 'setup'];
const VALID_ROUTES: Route[] = ['login', 'setup', 'dashboard', 'settings'];

function pathToRoute(path: string): Route {
  const clean = path.replace(/^\/+|\/+$/g, '') || 'dashboard';
  if (VALID_ROUTES.includes(clean as Route)) return clean as Route;
  return 'not-found';
}

function routeToPath(route: Route): string {
  return route === 'dashboard' ? '/' : `/${route}`;
}

export const currentRoute = writable<Route>(pathToRoute(window.location.pathname));
export const isPublicRoute = derived(currentRoute, ($route) =>
  PUBLIC_ROUTES.includes($route)
);

let isAuthenticated = false;

export function setAuthenticated(value: boolean): void {
  isAuthenticated = value;
}

export function navigate(route: Route): void {
  if (!isAuthenticated && !PUBLIC_ROUTES.includes(route)) {
    route = 'login';
  }
  const path = routeToPath(route);
  if (window.location.pathname !== path) {
    window.history.pushState({}, '', path);
  }
  currentRoute.set(route);
}

export function redirectToLogin(): void {
  navigate('login');
}

export function redirectToDashboard(): void {
  navigate('dashboard');
}

function handlePopState(): void {
  const route = pathToRoute(window.location.pathname);
  if (!isAuthenticated && !PUBLIC_ROUTES.includes(route)) {
    navigate('login');
    return;
  }
  currentRoute.set(route);
}

export function initRouter(): () => void {
  window.addEventListener('popstate', handlePopState);

  const route = get(currentRoute);
  if (!isAuthenticated && !PUBLIC_ROUTES.includes(route)) {
    navigate('login');
  }

  return () => {
    window.removeEventListener('popstate', handlePopState);
  };
}
