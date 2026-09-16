/**
 * Display label for an inbound: the remark when one is set, otherwise the
 * inbound tag. Falls back to an empty string when neither is present.
 * When a listener port is available, append it so identical remarks remain
 * distinguishable in client selectors.
 */
export function formatInboundLabel(tag?: string, remark?: string, port?: number): string {
  const remarkText = (remark || '').trim();
  const label = remarkText || (tag || '').trim();
  if (!label) return '';
  if (typeof port === 'number' && port > 0) return `${label} : ${port}`;
  return label;
}

export function formatTunnelConfigMeta(
  inbound: { id?: number; tag?: string; remark?: string },
  email?: string,
  totalCount = 1,
): {
  label?: string;
  fileName: string;
  qrRemark: string;
} {
  const inboundName =
    formatInboundLabel(inbound.tag, inbound.remark) ||
    (inbound.id != null ? `inbound-${inbound.id}` : '');
  const label = totalCount > 1 ? inboundName : undefined;
  const suffix = inbound.remark || inbound.tag || (inbound.id != null ? `${inbound.id}` : '');
  const safeSuffix = suffix ? `-${suffix.replace(/[^\w.-]+/g, '_')}` : '';
  const emailPrefix = email || 'client';
  const fileName = `${emailPrefix}${totalCount > 1 ? safeSuffix : ''}.conf`;
  const qrRemark =
    totalCount > 1 && inboundName ? [inboundName, email].filter(Boolean).join(' - ') : email || '';

  return { label, fileName, qrRemark };
}
