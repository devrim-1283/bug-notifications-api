import type { ReportFormData, ReportResponse, ErrorResponse } from './types';

export async function submitReport(
  data: ReportFormData,
  images: File[],
  turnstileToken: string
): Promise<ReportResponse> {
  const fd = new FormData();
  fd.append('site_id', data.siteId);
  fd.append('report_type', data.reportType);
  fd.append('title', data.title.trim());
  fd.append('category', data.category);
  fd.append('description', data.description.trim());

  if (data.pageUrl.trim()) fd.append('page_url', data.pageUrl.trim());
  if (data.fullName.trim()) fd.append('first_name', data.fullName.trim());
  if (data.email.trim()) {
    fd.append('contact_type', 'email');
    fd.append('contact_value', data.email.trim());
  } else if (data.phone.trim()) {
    fd.append('contact_type', 'phone');
    fd.append('contact_value', data.phone.trim());
  }

  images.forEach((f) => fd.append('images', f));
  fd.append('cf-turnstile-response', turnstileToken);

  const res = await fetch('/v1/reports', {
    method: 'POST',
    body: fd,
  });

  const body = await res.json();

  if (!res.ok) {
    const err = body as ErrorResponse;
    throw new Error(err.error || 'Submission failed');
  }

  return body as ReportResponse;
}
