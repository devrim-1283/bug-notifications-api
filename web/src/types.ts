export interface AppConfig {
  turnstileSiteKey: string;
  sites: string[];
  portalDomain: string;
}

export type ReportType = 'bug' | 'request';

export type Category =
  | 'design'
  | 'functionality'
  | 'performance'
  | 'content'
  | 'mobile'
  | 'security'
  | 'other';

export type Language = 'tr' | 'en' | 'de' | 'ru' | 'uk' | 'es';

export type AppView = 'form' | 'success';

export interface ReportFormData {
  siteId: string;
  reportType: ReportType;
  title: string;
  description: string;
  category: Category | '';
  pageUrl: string;
  fullName: string;
  phone: string;
  email: string;
}

export interface ReportResponse {
  event_id: string;
  queued: boolean;
}

export interface ErrorResponse {
  error: string;
  code: string;
}
