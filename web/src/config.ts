import type { AppConfig } from './types';

declare global {
  interface Window {
    __APP_CONFIG__: AppConfig | string;
  }
}

let cachedConfig: AppConfig | null = null;

const EMPTY_CONFIG: AppConfig = {
  apiKey: '',
  turnstileSiteKey: '',
  sites: [],
  portalDomain: '',
};

// Ensure all fields have safe defaults (guards against null from JSON)
function normalize(cfg: AppConfig): AppConfig {
  return {
    apiKey: cfg.apiKey ?? '',
    turnstileSiteKey: cfg.turnstileSiteKey ?? '',
    sites: cfg.sites ?? [],
    portalDomain: cfg.portalDomain ?? '',
  };
}

export function getConfig(): AppConfig {
  if (cachedConfig) return cachedConfig;

  const raw = window.__APP_CONFIG__;

  // Dev mode: placeholder not replaced
  if (!raw || raw === '__APP_CONFIG_JSON__') {
    cachedConfig = { ...EMPTY_CONFIG };
    return cachedConfig;
  }

  // Already an object (injected as raw JSON by Go server)
  if (typeof raw === 'object') {
    cachedConfig = normalize(raw as AppConfig);
    return cachedConfig;
  }

  // String fallback (shouldn't happen but safe)
  try {
    cachedConfig = normalize(JSON.parse(raw) as AppConfig);
  } catch {
    cachedConfig = { ...EMPTY_CONFIG };
  }

  return cachedConfig;
}
