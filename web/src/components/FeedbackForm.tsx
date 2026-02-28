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

export function FeedbackForm({ onSuccess }: Props) {
  const { t } = useI18n();
  const config = getConfig();

  const [form, setForm] = useState<ReportFormData>({ ...EMPTY_FORM });
  const [images, setImages] = useState<File[]>([]);
  const [turnstileToken, setTurnstileToken] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState('');
  const [autoDetectedSite, setAutoDetectedSite] = useState(false);

  useEffect(() => {
    const ref = document.referrer;
    if (!ref) return;
    try {
      const hostname = new URL(ref).hostname.toLowerCase();
      const match = config.sites.find(
        (s) => hostname === s || hostname.endsWith('.' + s)
      );
      if (match) {
        setForm((prev) => ({ ...prev, siteId: match }));
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
  }

  function handleReportTypeChange(type: ReportType) {
    if (type === form.reportType) return;
    const siteId = autoDetectedSite ? form.siteId : '';
    const pageUrl = autoDetectedSite ? form.pageUrl : '';
    setForm({ ...EMPTY_FORM, reportType: type, siteId, pageUrl });
    setImages([]);
    setTurnstileToken('');
    setError('');
  }

  const handleAutoFillUrl = useCallback((url: string) => {
    setForm((prev) => ({ ...prev, pageUrl: url }));
  }, []);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
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
            <SiteSelect
              value={form.siteId}
              onChange={(v) => updateField('siteId', v)}
              onAutoFillUrl={handleAutoFillUrl}
              disabled={autoDetectedSite}
              autoDetected={autoDetectedSite}
            />
            <div className="field">
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
            </div>
            <div className="field">
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
            </div>
          </div>

          {/* Right column */}
          <div className="form-col">
            <CategorySelect
              value={form.category}
              onChange={(v) => updateField('category', v as Category | '')}
            />
            <div className="field">
              <label>{t.labelPageUrl}</label>
              <input
                type="url"
                placeholder="https://..."
                value={form.pageUrl}
                onChange={(e) => updateField('pageUrl', e.target.value)}
              />
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
          onVerify={(token) => setTurnstileToken(token)}
          onExpire={() => setTurnstileToken('')}
          onError={() => setTurnstileToken('')}
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
