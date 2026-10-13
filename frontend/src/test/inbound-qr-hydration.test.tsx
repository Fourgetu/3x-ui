import { afterEach, describe, expect, it, vi } from 'vitest';
import { act, cleanup, fireEvent, screen } from '@testing-library/react';

import InboundsPage from '@/pages/inbounds/InboundsPage';
import type { InboundListProps } from '@/pages/inbounds/list/types';
import { DBInbound } from '@/models/dbinbound';
import fixture from './golden/fixtures/inbound-full/vless-tcp-reality.json';
import { renderWithProviders } from './test-utils';

const { hydrateInbound } = vi.hoisted(() => ({ hydrateInbound: vi.fn() }));
const slim = new DBInbound({
  ...fixture,
  nodeId: 7,
  settings: { ...fixture.settings, clients: [] },
});

vi.mock('@/layouts/AppSidebar', () => ({ default: () => null }));
vi.mock('@/hooks/useWebSocket', () => ({ useWebSocket: () => {} }));
vi.mock('@/api/queries/useNodesQuery', () => ({
  useNodesQuery: () => ({ nodes: [{ id: 7, address: 'node.example.test' }], fetched: true }),
}));
vi.mock('@/api/queries/useHostsQuery', () => ({
  useHostsQuery: () => ({ hosts: [], fetched: true }),
}));
vi.mock('@/pages/inbounds/useInbounds', () => ({
  useInbounds: () => ({
    fetched: true,
    dbInbounds: [slim],
    totals: { up: 0, down: 0 },
    subSettings: { enable: true },
    hydrateInbound,
  }),
}));
vi.mock('@/pages/inbounds/list', () => ({
  InboundList: ({ onRowAction }: InboundListProps) => (
    <button onClick={() => onRowAction({ key: 'qrcode', dbInbound: slim })}>Open inbound QR</button>
  ),
}));
vi.mock('@/pages/inbounds/qr/QrCodeModal', () => ({
  default: ({ clients, nodeAddress }: { clients: unknown[]; nodeAddress: string }) => (
    <output data-testid="qr-data">{JSON.stringify({ clients, nodeAddress })}</output>
  ),
}));

afterEach(() => {
  cleanup();
  vi.clearAllMocks();
});

describe('inbound QR detail loading', () => {
  it('loads client credentials before opening the QR from the slim list', async () => {
    hydrateInbound.mockResolvedValue(new DBInbound({ ...fixture, nodeId: 7 }));
    renderWithProviders(<InboundsPage />);
    fireEvent.click(screen.getByRole('button', { name: 'Open inbound QR' }));
    const payload = JSON.parse((await screen.findByTestId('qr-data')).textContent ?? '{}');
    expect(hydrateInbound).toHaveBeenCalledWith(fixture.id);
    expect(payload.clients[0].id).toBe(fixture.settings.clients[0].id);
    expect(payload.nodeAddress).toBe('node.example.test');
  });

  it('does not generate a credential-less QR when detail loading fails', async () => {
    hydrateInbound.mockResolvedValue(null);
    renderWithProviders(<InboundsPage />);
    await act(async () => {
      fireEvent.click(screen.getByRole('button', { name: 'Open inbound QR' }));
    });
    expect(hydrateInbound).toHaveBeenCalledWith(fixture.id);
    expect(screen.queryByTestId('qr-data')).toBeNull();
  });
});
