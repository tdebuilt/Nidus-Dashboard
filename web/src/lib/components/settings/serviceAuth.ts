/** Shared service authentication types and helpers for ServicesTab and AddServiceDialog. */

export interface ServiceTypeDef {
  type: string
  display_name: string
  auth_type: string
  needs_url: boolean
}

export function getServiceDef(
  defs: ServiceTypeDef[],
  type: string,
): ServiceTypeDef | undefined {
  return defs.find((d) => d.type === type)
}

export function serviceDisplayName(
  defs: ServiceTypeDef[],
  type: string,
): string {
  return (
    getServiceDef(defs, type)?.display_name ||
    type.charAt(0).toUpperCase() + type.slice(1)
  )
}

export function getAuthType(defs: ServiceTypeDef[], type: string): string {
  return getServiceDef(defs, type)?.auth_type || 'none'
}

export function needsURL(defs: ServiceTypeDef[], type: string): boolean {
  return getServiceDef(defs, type)?.needs_url ?? true
}

export function hasDualAuth(defs: ServiceTypeDef[], t: string): boolean {
  return getAuthType(defs, t) === 'dual'
}
export function hasTokenOnly(defs: ServiceTypeDef[], t: string): boolean {
  return getAuthType(defs, t) === 'token'
}
export function hasUserPassOnly(defs: ServiceTypeDef[], t: string): boolean {
  return getAuthType(defs, t) === 'userpass'
}
export function hasApiKeyOnly(defs: ServiceTypeDef[], t: string): boolean {
  return getAuthType(defs, t) === 'apikey'
}
export function hasPasswordOnly(defs: ServiceTypeDef[], t: string): boolean {
  return getAuthType(defs, t) === 'password'
}
export function hasJDAuth(defs: ServiceTypeDef[], t: string): boolean {
  return getAuthType(defs, t) === 'jdownloader'
}
export function hasNoAuth(defs: ServiceTypeDef[], t: string): boolean {
  return getAuthType(defs, t) === 'none'
}

export interface CredentialParams {
  serviceType: string
  authMode: 'token' | 'userpass'
  token: string
  username: string
  password: string
  proxmoxTokenId: string
  proxmoxTokenSecret: string
  jdEmail: string
  jdPassword: string
}

/** Build a JSON credentials string from form state. */
export function buildCredentials(
  defs: ServiceTypeDef[],
  params: CredentialParams,
): string {
  const authType = getAuthType(defs, params.serviceType)
  if (authType === 'dual') return buildDualAuthCredentials(params)

  const builders: Record<string, () => string> = {
    'none': () => '',
    'jdownloader': () => params.jdEmail
      ? JSON.stringify({ email: params.jdEmail, password: params.jdPassword })
      : '',
    'token': () => params.token ? JSON.stringify({ token: params.token }) : '',
    'userpass': () => params.username
      ? JSON.stringify({ username: params.username, password: params.password })
      : '',
    'apikey': () => params.token ? JSON.stringify({ api_key: params.token }) : '',
    'password': () => params.password
      ? JSON.stringify({ password: params.password })
      : '',
  }

  return (builders[authType] ?? builders['none'])()
}

function buildDualAuthCredentials(params: CredentialParams): string {
  if (params.authMode === 'token') {
    if (params.serviceType === 'proxmox') {
      if (!params.proxmoxTokenId || !params.proxmoxTokenSecret) return ''
      return JSON.stringify({
        token: `${params.proxmoxTokenId}=${params.proxmoxTokenSecret}`,
      })
    }
    return params.token ? JSON.stringify({ token: params.token }) : ''
  }
  return params.username
    ? JSON.stringify({ username: params.username, password: params.password })
    : ''
}
