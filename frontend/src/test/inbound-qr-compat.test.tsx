import { afterEach, describe, expect, it, vi } from 'vitest';
import { cleanup, fireEvent, render, screen } from '@testing-library/react';

import QrCodeModal from '@/pages/inbounds/qr/QrCodeModal';
import { buildRowActionsMenu } from '@/pages/inbounds/list/RowActions';
import type { DBInboundRecord } from '@/pages/inbounds/list/types';
import realityFixture from './golden/fixtures/inbound-full/vless-tcp-reality.json';

vi.mock('@/pages/inbounds/qr/QrPanel', () => ({
  default: ({ value, showQr }: { value: string; showQr: boolean }) => (
    <output data-testid="share-link" data-qr={String(showQr)}>
      {value}
    </output>
  ),
}));

afterEach(cleanup);

describe('inbound QR compatibility with upstream 3.10', () => {
  it('keeps the multi-client shortcut but omits protocols without share links', () => {
    const record: DBInboundRecord = {
      ...realityFixture,
      subSortIndex: 0,
      _expiryTime: null,
    };
    const keys = (protocol: string, hasClients: boolean) =>
      buildRowActionsMenu({
        record: { ...record, protocol },
        hasClients,
        subEnable: true,
        t: (k) => k,
      })?.map((item) => item?.key);
    expect(keys('vless', true)).toContain('qrcode');
    expect(keys('shadowsocks', true)).toContain('qrcode');
    expect(keys('vless', false)).not.toContain('qrcode');
    expect(keys('masque', true)).not.toContain('qrcode');
  });

  it('generates separately named client links on the hosting node and keeps REALITY QR visible', () => {
    const inbound = structuredClone(realityFixture);
    const first = inbound.settings.clients[0];
    const second = {
      ...first,
      id: '33333333-4444-4555-9666-777777777777',
      email: 'second@example.test',
    };
    const clients = [first, second];
    inbound.settings.clients = clients;
    render(
      <QrCodeModal
        open
        onClose={() => {}}
        dbInbound={inbound}
        clients={clients}
        protocolOnly
        nodeAddress="node.example.test"
      />,
    );
    const firstLink = screen.getByTestId('share-link');
    expect(firstLink.getAttribute('data-qr')).toBe('true');
    expect(firstLink.textContent).toContain(`${first.id}@node.example.test:443`);
    expect(decodeURIComponent(firstLink.textContent ?? '')).toContain(
      `${inbound.remark}-${first.email}`,
    );
    fireEvent.click(screen.getByText(`${inbound.remark}-${second.email}`));
    expect(
      screen.getAllByTestId('share-link').some((item) => item.textContent?.includes(second.id)),
    ).toBe(true);
  });

  it('uses the official oversized post-quantum key rule in the retained modal', () => {
    const inbound = structuredClone(realityFixture);
    inbound.streamSettings.realitySettings.settings.mldsa65Verify = 'B'.repeat(2603);
    render(
      <QrCodeModal
        open
        onClose={() => {}}
        dbInbound={inbound}
        clients={inbound.settings.clients}
        protocolOnly
      />,
    );
    expect(screen.getByTestId('share-link').getAttribute('data-qr')).toBe('false');
  });
});
