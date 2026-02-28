import { useState, useCallback, useEffect } from 'react';
import { useI18n } from '../i18n';
import { submitReport } from '../api';
import { getConfig } from '../config';
import type {
  ReportType,
  Category,
  ReportFormData,
} from '../types';
import { ReportTypeToggle } from './ReportTypeToggle';
import { SiteSelect } from './SiteSelect';
import { CategorySelect } from './CategorySelect';
import { ContactSection } from './ContactSection';
import { ImageUpload } from './ImageUpload';
import { TurnstileWidget } from './TurnstileWidget';

interface Props {
  onSuccess: () => void;
  resolvedTheme: 'light' | 'dark';
}

const EMPTY_FORM: ReportFormData = {
  siteId: '',
  reportType: 'bug',
  title: '',
  description: '',
  category: '',
  pageUrl: '',
  fullName: '',
  phone: '',
  email: '',
};

interface FieldErrors {
  siteId?: string;
  title?: string;
  category?: string;
  description?: string;
  pageUrl?: string;
}

function matchesSiteDomain(url: string, siteDomain: string): boolean {
  try {
    const hostname = new URL(url).hostname.toLowerCase();
    const domain = siteDomain.toLowerCase();
    return hostname === domain || hostname.endsWith('.' + domain);
  } catch {
    return false;
  }
}

export function FeedbackForm({ onSuccess, resolvedTheme }: Props) {
  const { t } = useI18n();
  const config = getConfig();

  const [form, setForm] = useState<ReportFormData>({ ...EMPTY_FORM });
  const [images, setImages] = useState<File[]>([]);
  const [turnstileToken, setTurnstileToken] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState('');
  const [autoDetectedSite, setAutoDetectedSite] = useState(false);
  const [fieldErrors, setFieldErrors] = useState<FieldErrors>({});

  useEffect(() => {
    const ref = document.referrer;
    if (!ref) return;
    try {
      const refUrl = new URL(ref);
      const hostname = refUrl.hostname.toLowerCase();
      const match = config.sites.find(
        (s) => hostname === s || hostname.endsWith('.' + s)
      );
      if (match) {
        setForm((prev) => ({ ...prev, siteId: match, pageUrl: ref }));
        setAutoDetectedSite(true);
      }
    } catch {
      // ignore
    }
  }, [config.sites]);

  function updateField<K extends keyof ReportFormData>(
    key: K,
    value: ReportFormData[K]
  ) {
    setForm((prev) => ({ ...prev, [key]: value }));
    // Clear field error when user starts typing/selecting
    if (key in fieldErrors) {
      setFieldErrors((prev) => {
        const next = { ...prev };
        delete next[key as keyof FieldErrors];
        return next;
      });
    }
  }

  function handleReportTypeChange(type: ReportType) {
    if (type === form.reportType) return;
    const siteId = autoDetectedSite ? form.siteId : '';
    const pageUrl = autoDetectedSite ? form.pageUrl : '';
    setForm({ ...EMPTY_FORM, reportType: type, siteId, pageUrl });
    setImages([]);
    setTurnstileToken('');
    setError('');
    setFieldErrors({});
  }

  const handleAutoFillUrl = useCallback((url: string) => {
    setForm((prev) => ({ ...prev, pageUrl: url }));
  }, []);

  function validate(): boolean {
    const errors: FieldErrors = {};

    if (!form.siteId) errors.siteId = t.errSiteRequired;
    if (!form.title.trim()) errors.title = t.errTitleRequired;
    if (!form.category) errors.category = t.errCategoryRequired;
    if (!form.description.trim()) errors.description = t.errDescRequired;
    if (form.pageUrl.trim() && form.siteId && !matchesSiteDomain(form.pageUrl.trim(), form.siteId)) {
      errors.pageUrl = t.errPageUrlDomain;
    }

    setFieldErrors(errors);
    return Object.keys(errors).length === 0;
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();

    if (!validate()) return;

    setError('');
    setSubmitting(true);

    try {
      await submitReport(form, images, turnstileToken);
      onSuccess();
    } catch (err) {
      setError(err instanceof Error ? err.message : t.errorGeneric);
      setTurnstileToken('');
    } finally {
      setSubmitting(false);
    }
  }

  const canSubmit = !submitting && !!turnstileToken;
  const isBug = form.reportType === 'bug';
  const titlePlaceholder = isBug ? t.titlePlaceholderBug : t.titlePlaceholderRequest;
  const descPlaceholder = isBug ? t.descPlaceholderBug : t.descPlaceholderRequest;

  return (
    <>
      <ReportTypeToggle
        value={form.reportType}
        onChange={handleReportTypeChange}
      />

      <form onSubmit={handleSubmit} noValidate>
        <div className="form-grid">
          {/* Left column */}
          <div className="form-col">
            <div className={`field${fieldErrors.siteId ? ' has-error' : ''}`}>
              <SiteSelect
                value={form.siteId}
                onChange={(v) => updateField('siteId', v)}
                onAutoFillUrl={handleAutoFillUrl}
                autoDetected={autoDetectedSite}
              />
              {fieldErrors.siteId && (
                <div className="field-error">{fieldErrors.siteId}</div>
              )}
            </div>
            <div className={`field${fieldErrors.title ? ' has-error' : ''}`}>
              <label>
                {t.labelTitle} <span className="req">*</span>
              </label>
              <input
                type="text"
                maxLength={200}
                required
                placeholder={titlePlaceholder}
                value={form.title}
                onChange={(e) => updateField('title', e.target.value)}
              />
              {fieldErrors.title && (
                <div className="field-error">{fieldErrors.title}</div>
              )}
            </div>
            <div className={`field${fieldErrors.description ? ' has-error' : ''}`}>
              <label>
                {t.labelDesc} <span className="req">*</span>
              </label>
              <textarea
                maxLength={5000}
                required
                placeholder={descPlaceholder}
                value={form.description}
                onChange={(e) => updateField('description', e.target.value)}
              />
              {fieldErrors.description && (
                <div className="field-error">{fieldErrors.description}</div>
              )}
            </div>
          </div>

          {/* Right column */}
          <div className="form-col">
            <div className={`field${fieldErrors.category ? ' has-error' : ''}`}>
              <CategorySelect
                value={form.category}
                onChange={(v) => updateField('category', v as Category | '')}
              />
              {fieldErrors.category && (
                <div className="field-error">{fieldErrors.category}</div>
              )}
            </div>
            <div className={`field${fieldErrors.pageUrl ? ' has-error' : ''}`}>
              <label>{t.labelPageUrl}</label>
              <input
                type="url"
                placeholder="https://..."
                value={form.pageUrl}
                onChange={(e) => updateField('pageUrl', e.target.value)}
              />
              {fieldErrors.pageUrl && (
                <div className="field-error">{fieldErrors.pageUrl}</div>
              )}
            </div>
            <ImageUpload files={images} onChange={setImages} />
          </div>
        </div>

        {/* Contact (full width) */}
        <ContactSection
          fullName={form.fullName}
          phone={form.phone}
          email={form.email}
          onFullNameChange={(v) => updateField('fullName', v)}
          onPhoneChange={(v) => updateField('phone', v)}
          onEmailChange={(v) => updateField('email', v)}
        />

        {/* Turnstile */}
        <TurnstileWidget
          onVerify={useCallback((token: string) => setTurnstileToken(token), [])}
          onExpire={useCallback(() => setTurnstileToken(''), [])}
          onError={useCallback(() => setTurnstileToken(''), [])}
          theme={resolvedTheme}
        />

        {/* Submit */}
        <button type="submit" className="btn-submit" disabled={!canSubmit}>
          {submitting ? (
            <>
              <span className="spinner" /> {t.sending}
            </>
          ) : (
            <>
              <i className="fa-solid fa-paper-plane" /> {t.submitBtn}
            </>
          )}
        </button>

        {error && <div className="msg error">{error}</div>}
      </form>
    </>
  );
}
