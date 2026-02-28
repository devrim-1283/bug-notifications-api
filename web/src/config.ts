import type { AppConfig } from './types';

declare global {
  interface Window {
    __APP_CONFIG__: string;
  }
}

let cachedConfig: AppConfig | null = null;

export function getConfig(): AppConfig {
  if (cachedConfig) return cachedConfig;

  const raw = window.__APP_CONFIG__;
  if (!raw || raw === '__APP_CONFIG_JSON__') {
    // Dev fallback
    cachedConfig = {
      apiKey: '',
      turnstileSiteKey: '',
      sites: [],
      portalDomain: '',
    };
    return cachedConfig;
  }

  try {
    cachedConfig = JSON.parse(raw) as AppConfig;
  } catch {
    cachedConfig = {
      apiKey: '',
      turnstileSiteKey: '',
      sites: [],
      portalDomain: '',
    };
  }

  return cachedConfig;
}
