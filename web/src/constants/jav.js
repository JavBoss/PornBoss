export const JAV_PROVIDER_UNKNOWN = 0
export const JAV_PROVIDER_JAVBUS = 1
export const JAV_PROVIDER_JAVDATABASE = 2
export const JAV_PROVIDER_USER = 3

export function isUserJavTag(tag) {
  return Number(tag?.provider) === JAV_PROVIDER_USER
}
